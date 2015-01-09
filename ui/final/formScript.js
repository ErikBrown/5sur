var form = document.getElementById('plain_form');

function setFormHeight() {
	form.style.minHeight = window.innerHeight - 100 + "px";
}

window.onload = function() {
	setFormHeight();
}

window.onresize = function() {
	setFormHeight();
}