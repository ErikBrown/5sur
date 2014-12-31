/***** Form Page *****/
var sidebar1 = document.getElementById('sidebar');
var sidebar2 = document.getElementById('sub_sidebar');
var header = document.getElementById('header');
var header2 = document.getElementById('content_header');

function setSidebarHeight() {
	sidebar1.style.minHeight = window.innerHeight - header.offsetHeight - header2.offsetHeight + "px";
	sidebar2.style.minHeight = window.innerHeight - header.offsetHeight - header2.offsetHeight + "px";
}

window.onload = function() {
	setSidebarHeight();
}

window.onresize = function() {
	setSidebarHeight();
}