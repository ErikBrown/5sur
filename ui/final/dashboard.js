var sidebar1 = document.getElementById('main_sidebar');
var sidebar2 = document.getElementById('sub_sidebar');
var header = document.getElementById('header');

function setSidebarHeight() {
	var body = document.body,
		html = document.documentElement;

	var height = Math.max(body.scrollHeight, body.offsetHeight, 
				html.clientHeight, html.scrollHeight, html.offsetHeight);
	sidebar1.style.height = height - header.offsetHeight + "px";
	sidebar2.style.height = height - header.offsetHeight + "px";
	console.log(height - header.offsetHeight + "px");
}

window.onload = function() {
	setSidebarHeight();
}

window.onresize = function() {
	setSidebarHeight();
}
