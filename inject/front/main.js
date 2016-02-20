// TODO: add columns so that they are alphabetically sorted
function addColumn(name) {
  // append new column header
  $("#main-table tr:first").append("<td id='fn_" + name + "'>" + name + "</td>");

  var rows = $("#main-table tr");
  for (var i = 1; i < $("#main-table tr").length - 1; i++) {
    var rowId = rows[i].id;
    var [chan, fn1, fn2] = rowId.split("/");
    if (name == fn1 || name == fn2) {
      // add the name at the end to know under each function is this cell id
      var tdId = chan + "_" + fn1 + "_" + fn2 + "_" + name;
      $("#main-table tr:eq(" + i + ")").append("<td id='" + tdId + "'>o</td>");
    } else {
      $("#main-table tr:eq(" + i + ")").append("<td></td>");
    }
  }
}

// NOTE: cells have id of the following pattern:
// <chan>_fn_<ithFn>_fn_<kthFn>_fn_<columnFn>

function addRow(name) {
  var headerRow = $("#main-table tr:first")[0].children;
  var columns = headerRow.length

  for (var i = 1; i < columns; i++) {
    for (var k = i; k < columns; k++) {
      var ithFn = headerRow[i].id;
      var kthFn = headerRow[k].id;
      var trId = name + "_" + ithFn + "_" + kthFn;

      $("#main-table tbody").append("<tr id='" + trId + "'></tr>");

      var newRow = $("#main-table tbody tr:last");

      newRow.append("<td class='chan-name'>" + name + "</td>");

      for (var l = 1; l < headerRow.length; l++) {
        if (headerRow[l].id == ithFn || headerRow[l].id == kthFn) {
          var tdId = trId + "_" + headerRow[l].id;
          newRow.append("<td id='" + tdId + "'>o</td>");
        } else {
          newRow.append("<td></td>");
        }
      }
    }
  }

  $(".chan-name").outerWidth(100);
}

function idFromSingleMessage(message) {
  return message.Chan + ("_fn_" + message.Func).repeat(3);
}

function processRecive(recieveMessage) {
  var id = idFromSingleMessage(recieveMessage);

  document.getElementById(id).innerHTML = "<-o";
}

function processSend(sendMessage) {
  var id = idFromSingleMessage(sendMessage);

  document.getElementById(id).innerHTML = "o<-";
}

function processMessagePassing(sendMessage, recieveMessage) {
  var chan = sendMessage.Chan;
  var sendFn = sendMessage.Func;
  var recvFn = recieveMessage.Func;

  var leftColumn, rightColumn;
  var testId = chan + "_fn_" + sendFn + "_fn_" + recvFn + "_fn_" + sendFn;
  if (document.getElementById(testId) != null) {
    leftColumn = sendFn;
    rightColumn = recvFn;
  } else {
    leftColumn = recvFn;
    rightColumn = sendFn;
  }

  var common = chan + "_fn_" + leftColumn + "_fn_" + rightColumn + "_fn_";
  var leftId = common + leftColumn;
  var rightId = common + rightColumn;
  if (leftColumn == sendMessage.Func) {
    document.getElementById(leftId).innerHTML = "o----->" + sendMessage.Value;
    document.getElementById(rightId).innerHTML = "----->o";
  } else {
    document.getElementById(leftId).innerHTML = "o<-----";
    document.getElementById(rightId).innerHTML = sendMessage.Value + "<-----o";
  }
}

var events = {};
var channels = {};

$( document ).ready(function() {
  var connection = new WebSocket('ws://127.0.0.1:9610/');

  connection.onopen = function() {
    connection.onmessage = function(e) {
      var serverMessage = JSON.parse(e.data);

      // if the message is the functions initialization message
      if ('Functions' in serverMessage) {
        for (i in serverMessage.Functions) {
          addColumn(serverMessage.Functions[i]);
        }

        return;
      }

      // if it is a normal event message

      if (!(serverMessage.Chan in channels)) {
        channels[serverMessage.Chan] = true;
        addRow(serverMessage.Chan);
      }

      // if it is a read event
      if (serverMessage.Type == true) {
        // the key for a corresponding send event
        var key = serverMessage.Chan + '.false.' + serverMessage.Value;
        if (key in events) {
          var correspondingSend = events[key].shift();

          document.getElementById(idFromSingleMessage(correspondingSend)).innerHTML = "o";

          if (events[key].length == 0) {
            delete events[key];
          }

          processMessagePassing(correspondingSend, serverMessage);
        } else {
          var recieveEventKey = serverMessage.Chan + '.true.' + serverMessage.Value;

          if (recieveEventKey in events) {
            events[recieveEventKey].push(serverMessage);

            processRecive(serverMessage);
          } else {
            events[recieveEventKey] = [serverMessage];

            processRecive(serverMessage);
          }
        }
      } else { // if it is a write event
        // the key for a corresponding recieve event
        var key = serverMessage.Chan + '.true.' + serverMessage.Value;
        if (key in events) {
          var correspondingRecieve = events[key].shift();

          document.getElementById(idFromSingleMessage(correspondingRecieve)).innerHTML = "o";

          if (events[key].length == 0) {
            delete events[key];
          }

          processMessagePassing(serverMessage, correspondingRecieve);
        } else {
          var sendEventKey = serverMessage.Chan + '.false.' + serverMessage.Value;

          if (sendEventKey in events) {
            events[sendEventKey].push(serverMessage);

            processSend(serverMessage);
          } else {
            events[sendEventKey] = [serverMessage];

            processSend(serverMessage);
          }
        }
      }
    }
  }
});
