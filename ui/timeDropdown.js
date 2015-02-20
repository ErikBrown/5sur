/***** Calendar Widget *****/
// http://stackoverflow.com/a/901144/2547044
function getParameterByName(name) {
	name = name.replace(/[\[]/, "\\[").replace(/[\]]/, "\\]");
	var regex = new RegExp("[\\?&]" + name + "=([^&#]*)"),
		results = regex.exec(location.search);
	return results === null ? "" : decodeURIComponent(results[1].replace(/\+/g, " "));
}

function buildDropdown(h) {
	var h = getParameterByName("h");
	var select = document.getElementById("time_box")
	var option = document.createElement('option');
	for(i=0; i<48; i++) {
		temp = option.cloneNode(true);
		var j = i/2;
		hour = ""
		if (j === parseInt(j, 10)) {
			hour = j + ":00"
		} else {
			hour = j - .5 + ":30"
		}
		if (j < 10) {
			hour = "0" + hour;
		}
		temp2 = document.createTextNode(hour);
		temp.appendChild(temp2);
		if (hour == h) {
			temp.selected = "true";
		}
		select.appendChild(temp);
	}
}

window.addEventListener("load", buildDropdown, false);