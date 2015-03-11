function adjustHeight() {
		document.getElementById("sub_sidebar").style.height = "0px";
		document.getElementById("main_sidebar").style.height = "0px";
	var body = document.body,
	html = document.documentElement;

	var h = Math.max( body.scrollHeight, body.offsetHeight, 
		html.clientHeight, html.scrollHeight, html.offsetHeight );
	var w = window.innerWidth;
	if (w > 1050) {
		document.getElementById("main_sidebar").style.height = h - 75 + "px";
		document.getElementById("sub_sidebar").style.height = h - 75 + "px";
	} else {
		document.getElementById("main_sidebar").style.height = "250px";
		document.getElementById("sub_sidebar").style.height = h - 325 + "px";
	}
}

window.addEventListener("load", function(){
	adjustHeight();
}, false);

window.addEventListener("resize", function(){
	adjustHeight();
}, false);