// TODO: add columns so that they are alphabetically sorted
function addColumn(name) {
  // append new column header
  $("#main-table tr:first").append("<td id='fn@" + name + "'>" + name + "</td>");

  var rows = $("#main-table tr");
  for (var i = 1; i < $("#main-table tr").length - 1; i++) {
    var rowId = rows[i].id;
    var [chan, fn1, fn2] = rowId.split("#");
    if (name == fn1 || name == fn2) {
      // add the name at the end to know under each function is this cell id
      var tdId = chan + "#" + fn1 + "#" + fn2 + "#" + name;
      $("#main-table tr:eq(" + i + ")").append("<td id='" + tdId + "'>o</td>");
    } else {
      $("#main-table tr:eq(" + i + ")").append("<td></td>");
    }
  }
}

function addRow(name) {
  var headerRow = $("#main-table tr:first")[0].children;
  var columns = headerRow.length

  for (var i = 1; i < columns; i++) {
    for (var k = i; k < columns; k++) {
      var ithFn = headerRow[i].id;
      var kthFn = headerRow[k].id;
      var trId = name + "#" + ithFn + "#" + kthFn;

      $("#main-table tbody").append("<tr id='" + trId + "'></tr>");

      var newRow = $("#main-table tbody tr:last");

      newRow.append("<td>" + name + "</td>");

      for (var l = 1; l < headerRow.length; l++) {
        if (headerRow[l].id == ithFn || headerRow[l].id == kthFn) {
          var tdId = trId + "#" + headerRow[l].id;
          newRow.append("<td id='" + tdId + "'>o</td>");
        } else {
          newRow.append("<td></td>");
        }
      }
    }
  }
}

function processRecive(recieveMessage) {
  console.log("proc", recieveMessage);
}

function processSend(sendMessage) {
  console.log("proc", sendMessage);
}

function processMessagePassing(sendMessage, recieveMessage) {
  console.log(sendMessage + " -> " + recieveMessage);
}

var events = {};
var channels = {};

$( document ).ready(function() {
  var connection = new WebSocket('ws://127.0.0.1:9610/');

  connection.onopen = function() {
    connection.onmessage = function(e) {
      var serverMessage = JSON.parse(e.data);
      console.log(serverMessage);

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
        var key = serverMessage.Chan + '.' + serverMessage.Value + ".false";
        if (key in events) {
          var correspondingSend = events[key].shift();

          if (events[key].length == 0) {
            delete events[key];
          }

          processMessagePassing(correspondingSend, serverMessage);
        } else {
          var recieveEventKey = serverMessage.Chan + '.' + serverMessage.Value + ".true";

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
        var key = serverMessage.Chan + '.' + serverMessage.Value + ".true";
        if (key in events) {
          var correspondingRecieve = events[key].shift();

          if (events[key].length == 0) {
            delete events[key];
          }

          processMessagePassing(serverMessage, correspondingRecieve);
        } else {
          var sendEventKey = serverMessage.Chan + '.' + serverMessage.Value + ".true";

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
