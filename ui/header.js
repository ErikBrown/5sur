/***** Form Page *****/
var user = document.getElementById('nav_user');
var userDropdown = document.getElementById('nav_user_dropdown');
var alert = document.getElementById('nav_alert');
var alertDropdown = document.getElementById('nav_alert_dropdown');

function getEventTarget(e) {
	e = e || window.event;
	return e.target || e.srcElement; 
}

document.addEventListener('click', function(event) {
	userDropdown.style.height = '0';
	alertDropdown.style.height = '0';
}, false);

user.addEventListener('click', function(event) {
	alertDropdown.style.height = '0';
	userDropdown.style.height = '158px';
	event.stopPropagation();
}, false);

userDropdown.addEventListener('click', function(event) {
	event.stopPropagation();
}, false);

alert.addEventListener('click', function(event) {
	userDropdown.style.height = '0';
	alertDropdown.style.height = '300px';
	event.stopPropagation();
}, false);

alertDropdown.addEventListener('click', function(event) {
	event.stopPropagation();
}, false);

/*
function setFormHeight() {
	form.style.minHeight = window.innerHeight - 100 + "px";
}

window.onload = function() {
	setFormHeight();
}
*/