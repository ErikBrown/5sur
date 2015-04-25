/** Global variables **/
var canvas = document.getElementById("donut_bars1");
var canvas2 = document.getElementById("donut_text1");
var canvas2_2 = document.getElementById("donut_text1_2");
var canvas3 = document.getElementById("donut_bars2");
var canvas4 = document.getElementById("donut_text2");
var canvas4_2 = document.getElementById("donut_text2_2");
var canvas5 = document.getElementById("line_graph");
var canvas5_2 = document.getElementById("line_graph_text");
var c = ["#B0F9F9", "#fbc93d", "#fb4f4f","#63C76A"]; // add more years!
var graphWidth = 225;

/****** Donut charts ******/
function createDonutChart(ctx, ctx2, ctx2_2, v, text) {
	var inc = 1;
	var vTotal = dataTotal(v);
	var vProportion = dataToRadians(v, vTotal);
	var label = labelPosition(ctx, vProportion);
	initialRotate(ctx);
	var donutChartInterval;
	donutChartInterval = setInterval(function() {
		inc = animateDonutChart(ctx, ctx2, ctx2_2, vProportion, vTotal, label, text, inc, donutChartInterval);
	}, 15);
}

// Sum of the data
function dataTotal(v) {
	var vT = 0;
	for(var i=0, len=v.length; i<len; i++) {
		vT += v[i];
	}
	return vT;
}

// Data represented proportionally as radians
function dataToRadians(v, vT) {
	var v2 = [];
	for(var i=0, len=v.length; i<len; i++) {
		if (i !=0) {
			v2[i] = v2[i-1] + (v[i]/vT * 2);
		} else {
			v2[i] = v[i]/vT * 2;
		}
	}
	return v2
}

// Returns object with x,y coordinates of donut chart label
function labelPosition(ctx, v) {
	ctx.font="300 14px Open Sans, sans-serif";
	var l = new Object();
	l.x = [0,0,0];
	l.y = [];
	for(var i=0, len=v.length; i<len; i++) {
		var radPos = 0;
		if (i == 0) {
			radPos = 2-(v[0]/2);
		} else {
			radPos = 2-(v[i]/2+v[i-1]/2);
		}
		l.x[i] = (graphWidth/2+Math.cos(radPos*Math.PI)*70);
		l.x[i] -= ctx.measureText(yearCreated).width/2;
		l.y[i] = (graphWidth/2+Math.sin(radPos*Math.PI)*70);
		l.y[i] += 7; // Font height offset

		var radPosTemp = 0;
		// Label offset
		l.x[i] += Math.cos(radPos*Math.PI) * 23;
		l.y[i] += Math.sin(radPos*Math.PI) * 23;
	}
	return l;
}

// Rotates graph .1pi radians before chart animation
function initialRotate(ctx) {
	ctx.translate(graphWidth/2, graphWidth/2);
	ctx.rotate(-Math.PI*.1);
	ctx.translate(-graphWidth/2, -graphWidth/2);
}

// Run once at window.onload, iterates 100 times
function animateDonutChart(ctx, ctx2, ctx2_2, v, vT, l, text, inc, chartInt) {
	ctx.clearRect(0,0,graphWidth,graphWidth);
	ctx2.clearRect(0,0,graphWidth,graphWidth);
	ctx.translate(graphWidth/2, graphWidth/2);
	ctx.rotate(Math.PI*.001);
	ctx.translate(-graphWidth/2, -graphWidth/2);
	if (inc != 100) {
		donutSlice(ctx, Math.PI*2, c[0], inc);
	} else {
		donutSlice(ctx, 0, c[0], inc);
	}
	for(i=v.length-2; i>=0; i--) {
		if (v[i] != 0) {
			if (inc == 100 && v[i] == 2) {
				donutSlice(ctx, 0, c[i+1], inc);
			} else {
				donutSlice(ctx, Math.PI*v[i], c[i+1], inc);
			}
		}
	}

	// inner circle
	ctx.beginPath();
	ctx.arc(graphWidth/2,graphWidth/2,67,0,Math.PI*2,true); 
	ctx.fillStyle = "#3A5A6D";
	ctx.fill();

	// border
	ctx.beginPath();
	ctx.arc(graphWidth/2,graphWidth/2,70,0,Math.PI*2,true);
	ctx.strokeStyle = "#3A5A6D";
	ctx.lineWidth = 2;
	ctx.stroke();

	ctx2.fillStyle ="rgba(255,255,255,.8)"; // Make fade-in function
	ctx2.font="18px Montserrat";
	ctx2.fillText(parseInt(vT*(inc/100)),(graphWidth/2)-ctx2.measureText(vT).width/2,graphWidth/2);

	ctx2.fillStyle ="rgba(255,255,255,"  + inc/160 + ")"; // Make fade-in function
	ctx2.font="14px Open Sans";
	ctx2.fillText(text,(graphWidth/2)-ctx2.measureText(text).width/2,graphWidth/2+20);

	if (inc >= 100) {
		for(var i=0, len=v.length; i<len; i++) {
			ctx2_2.fillStyle ="rgba(255,255,255," + inc/100 + ")"; // Make fade-in function
			ctx2_2.font="300 14px Open Sans, sans-serif";
			ctx2_2.fillText(yearCreated + i,l.x[i],[l.y[i]]);
		}
		clearInterval(chartInt);
	}

	inc = inc + 1.5;
	return inc;
}

// Is fired v.length times each graphInterval
function donutSlice(ctx, r, c, inc) {
	ctx.beginPath();
	ctx.arc(graphWidth/2,graphWidth/2,70,0,(Math.PI*2)-(r*(inc/100)),true);
	ctx.lineTo(graphWidth/2,graphWidth/2);
	ctx.fillStyle = c;
	ctx.fill();
}

window.onresize = function() {
	if (canvas.getContext) {
		// resizeLineGraph(false);
	} else {
		// Browser is not supported
	}

}

window.onload = function() {
	var ctx = canvas.getContext("2d");
	var ctx2 = canvas2.getContext("2d");
	var ctx2_2 = canvas2_2.getContext("2d");
	var ctx3 = canvas3.getContext("2d");
	var ctx4 = canvas4.getContext("2d");
	var ctx4_2 = canvas4_2.getContext("2d");
	if (canvas.getContext) {
		createDonutChart(ctx, ctx2, ctx2_2, ridesGiven, "Viajes compartidas");
		createDonutChart(ctx3, ctx4, ctx4_2, ridesTaken, "Viajes tomadas");
	} else {
		// Browser is not supported
	}
}