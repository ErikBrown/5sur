var form = document.getElementById('login_form');




function setFormHeight() {
	form.style.minHeight = window.innerHeight - 100 + "px";
}

window.onload = function() {
	setFormHeight();
}

window.onresize = function() {
	setFormHeight();
}