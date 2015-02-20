/***** Calendar Widget *****/

var today = new Date();
var months = ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'];
var monthColors = ['#35A6D0', '#D37056', '#61BE71', '#7C5498', '#AE75A3', '#54A7CF', '#60AE6D', '#C4BF66', '#D9B340', '#71ACB3', '#947C52', '#2D4B93'];


var calendar = document.getElementById('calendar');
var calendarWrapper = document.getElementById('calendar_wrapper');
var dateInput = document.getElementById('date_box');
var monthRight = document.getElementById('month_right');
var monthLeft = document.getElementById('month_left');
var hideCalendar = false;

function createTable(d, n) {
	var currentMonth = d.getMonth();
	d.setDate(1);
	var temp = d.getDay();
	if (d.getDay() == 0) {
		temp = 6;
	}
	d.setDate(1 - temp);

	var monthWrapper = document.getElementById('month_wrapper');
	monthWrapper.style.background = monthColors[currentMonth];
	
	var calendarMonth = document.getElementById('month_title');
	var monthText = document.createTextNode(months[currentMonth]);
	
	calendarMonth.innerHTML = "";
	calendarMonth.appendChild(monthText);
	
	var tableBody = document.createElement('tbody'); 
	var tableRowRef = document.createElement('tr');
	var tableCellRef = document.createElement('td');
	for (var i = 0; i < 6; i++) {
		tableRow = tableRowRef.cloneNode(true);
		for (var j = 0; j < 7; j++) {
			d.setDate(d.getDate() + 1);
			tableCell = tableCellRef.cloneNode(true);
			if (d.getMonth() != currentMonth) {
				var cellText = document.createTextNode(d.getDate());
				tableCell.className = d.getMonth() + ' ' + d.getFullYear() + ' other_month';
				tableCell.appendChild(cellText);
			} else {
				var cellText = document.createTextNode(d.getDate());
				tableCell.className = d.getMonth() + ' ' + d.getFullYear();
				tableCell.appendChild(cellText);
			}
			tableRow.appendChild(tableCell);
		}
		tableBody.appendChild(tableRow);
	}
	var tbodyNum = document.getElementsByTagName("tbody");
	if (tbodyNum.length != 0) {
		if (tbodyNum.length == 2) {	
			tbodyNum[0].parentNode.removeChild(tbodyNum[0]);
		}
		if (n == true) {
			tableBody.style.left = calendar.offsetWidth + 'px';
		} else {
			tableBody.style.left = '-' + calendar.offsetWidth + 'px';
		}
	} else {
		calendarWrapper.style.width = calendar.offsetWidth + 'px';	
	}
	tableBody.style.width = calendar.offsetWidth + 'px';	
	calendar.appendChild(tableBody);
}

monthRight.addEventListener('click', function() {
	createTable(today, true);
	var tBody = calendar.getElementsByTagName('tbody');
	tBody[1].style.borderLeft = "1px solid #cccccc";
	tBody[0].style.left = '-' + calendar.offsetWidth + 'px';
	tBody[1].style.left = '0';
	setTimeout(function(){
		tBody[1].style.borderLeft = "0px";
	}, 750);
}, false);

monthLeft.addEventListener('click', function() {
	today.setDate(today.getDate() - 42);
	createTable(today, false);
	var tBody = calendar.getElementsByTagName('tbody');
	tBody[0].style.borderLeft = "1px solid #cccccc";
	tBody[0].style.left = calendar.offsetWidth + 'px';
	tBody[1].style.left = '0';
	setTimeout(function(){
		tBody[0].style.borderLeft = "0px";
	}, 750);
}, false);

dateInput.addEventListener('click', function() {
	calendarWrapper.style.display = 'block';
	calendarWrapper.style.opacity = '1';
	calendarWrapper.style.height = '505px';
	calendarWrapper.style.marginBottom = '25px';
	hideCalendar = false;
}, false);

function getEventTarget(e) {
	e = e || window.event;
	return e.target || e.srcElement; 
}

calendarWrapper.addEventListener('click', function(event) {
	var calendarItem = getEventTarget(event);
	var calendarMonth = document.getElementById('month_title');
	if (calendarItem.nodeName == 'TD') {
		var tempDate = calendarItem.className.split(' ');
		dateInput.value =  calendarItem.innerHTML + '/' + (parseInt(tempDate[0]) + 1) + '/' + parseInt(tempDate[1]);
		calendarWrapper.style.height = '0px';
		calendarWrapper.style.opacity = '.8';
		calendarWrapper.style.marginBottom = '0px';
	} else {
		hideCalendar = false;
	}
}, false);


document.addEventListener('click', function() {
	if (hideCalendar == true) {
		calendarWrapper.style.height = '0px';
		calendarWrapper.style.opacity = '.8';
		calendarWrapper.style.marginBottom = '0px';
	}
	hideCalendar = true;
}, false);

window.addEventListener("load", function(){
	createTable(today, true);
}, false);