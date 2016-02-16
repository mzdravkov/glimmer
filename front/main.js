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

  for (var i = 1; i < columns - 1; i++) {
    for (var k = i + 1; k < columns; k++) {
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

$( document ).ready(function() {
  $.getJSON("glimmer_functions.json", function(data) {
    alert(data);
  });
});
