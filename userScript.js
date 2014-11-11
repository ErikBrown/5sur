/** Global variables **/
var canvas = document.getElementById("donut_bars1");
var canvas2 = document.getElementById("donut_text1");
var canvas2_2 = document.getElementById("donut_text1_2");
var canvas3 = document.getElementById("donut_bars2");
var canvas4 = document.getElementById("donut_text2");
var canvas4_2 = document.getElementById("donut_text2_2");
var canvas5 = document.getElementById("line_graph");
var canvas5_2 = document.getElementById("line_graph_text");
var c = ["#6cc0e5", "#fbc93d", "#fb4f4f","#18B64B"];
var yearCreated = 2014;

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
		l.x[i] = (300/2+Math.cos(radPos*Math.PI)*90);
		l.x[i] -= ctx.measureText(yearCreated).width/2;
		l.y[i] = (300/2+Math.sin(radPos*Math.PI)*90);
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
	ctx.translate(300/2, 300/2);
	ctx.rotate(-Math.PI*.1);
	ctx.translate(-300/2, -300/2);
}

// Run once at window.onload, iterates 100 times
function animateDonutChart(ctx, ctx2, ctx2_2, v, vT, l, text, inc, chartInt) {
	// console.log((intervalFactor*.01).toPrecision(2));
	ctx.clearRect(0,0,300,300);
	ctx2.clearRect(0,0,300,300);
	ctx.translate(300/2, 300/2);
	ctx.rotate(Math.PI*.001);
	ctx.translate(-300/2, -300/2);
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
	ctx.arc(300/2,300/2,87,0,Math.PI*2,true); 
	ctx.fillStyle = "#F3F3F3";
	ctx.fill();

	// border
	ctx.beginPath();
	ctx.arc(300/2,300/2,90,0,Math.PI*2,true);
	ctx.strokeStyle = "#F3F3F3";
	ctx.lineWidth = 2;
	ctx.stroke();

	ctx2.fillStyle ="rgba(80,80,80,1)"; // Make fade-in function
	ctx2.font="21px Montserrat";
	ctx2.fillText(parseInt(vT*(inc/100)),(300/2)-ctx2.measureText(vT).width/2,300/2);

	ctx2.fillStyle ="rgba(125,125,125,"  + inc/100 + ")"; // Make fade-in function
	ctx2.font="16px Montserrat";
	ctx2.fillText(text,(300/2)-ctx2.measureText(text).width/2,300/2+20);

	if (inc >= 100) {
		for(var i=0, len=v.length; i<len; i++) {
			ctx2_2.fillStyle ="rgba(150,150,150," + inc/100 + ")"; // Make fade-in function
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
	ctx.arc(300/2,300/2,90,0,(Math.PI*2)-(r*(inc/100)),true);
	ctx.lineTo(300/2,300/2);
	ctx.fillStyle = c;
	ctx.fill();
}

/****** Line Graph ******/
 function resizeLineGraph(animate) {
	width = document.getElementById("user_graph").offsetWidth;
	canvas5.width = width;
	canvas5_2.width = width;
	var inc = 1;
	if (animate == false) {
		drawLineGraph(width, lineGraphInterval, 100);
	} else {
	lineGraphInterval = setInterval(function() {
		inc = drawLineGraph(width, lineGraphInterval, inc);
	}, 15);

	}
}

// THIS IS ALL TEMP
var monthValue = [10,7,13,8,7,12,11,12,9,10,6,8];
var monthHighest = 0;

for (i=monthValue.length-1; i >= 0; i--) {
	if (monthHighest < monthValue[i]) {
		monthHighest = monthValue[i];
	}
}

var monthProportion = [];

for (i=0, len = monthValue.length-1; i <= len; i++) {
	monthProportion[monthValue.length-1-i] = -220 * (monthValue[i]/monthHighest);
}

var monthName = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

// Temp
var circleX = [];
var circleY = [];

function drawLineGraph(w, graphInt, inc) {
	var ctx = canvas5.getContext("2d");
	var ctx2 = canvas5_2.getContext("2d");
	ctx.clearRect(0,0,300,300);
	ctx.translate(w, 250);

	// Graph Line
	ctx.beginPath();
	var x = 0;
	for (i=0, len=monthProportion.length; i < len; i++) {
		// createBarGraphLabels(ctx2, -x, monthName[i]);
		x = fillGraph(ctx, w, x, monthProportion[i]);	
	}
	ctx.lineTo(x,0);
	ctx.lineTo(0,0);
	ctx.fillStyle = "#C8ECF0";
	ctx.fill();

	ctx.strokeStyle = "#7DD1DD";
	ctx.lineWidth = 2;
	ctx.stroke();

	// Vertical Lines
	ctx.closePath();
	ctx.beginPath();
	x = 0;
	for (i=0, len=monthProportion.length; i < len; i++) {
		x = verticalLine(ctx, w, x, monthProportion[i], inc);	
	}
	ctx.strokeStyle = "#ABE1E9";
	ctx.lineWidth = 1;
	ctx.stroke();

	// Point Circles
	ctx.closePath();
	ctx.beginPath();
	x = 0;
	for (i=0, len=monthProportion.length; i < len; i++) {
		circleX[i] = x;
		circleY.unshift(monthProportion[i]);
		x = pointCircle(ctx, w, x, monthProportion[i], inc);
	}
	ctx.fillStyle = "#ffffff";
	ctx.fill();

	ctx.strokeStyle = "#7DD1DD";
	ctx.lineWidth = 2;
	ctx.stroke();
	if (inc >= 100) {
		clearInterval(graphInt);
	}
	return inc + 1.5;
}

function fillGraph(ctx, w, x, y) {
	ctx.lineTo(x,y);
	return x-(w/11);
}

function verticalLine(ctx, w, x, y, inc) {
	ctx.moveTo(x,y);
	ctx.lineTo(x,0);
	return x-(w/11);
}

function pointCircle(ctx, w, x, y, inc) {
	ctx.moveTo(x+5,y);
	ctx.arc(x,y,5,0,Math.PI*2,true); 
	return x-(w/11);
}

function getCursorPos(e) {
	return {
		x: e.clientX - canvas5.getBoundingClientRect().left,
		y: e.clientY - canvas5.getBoundingClientRect().top
	};
}

function plotPointDetails(ctx2, x, y, i) {
	ctx2.fillStyle ="rgba(160,160,160,1)"; // Make fade-in function
	ctx2.font="16px Open Sans";
	ctx2.fillText(monthValue[i],-x-ctx2.measureText(monthValue[i]).width/2,240+y-5);
}

window.addEventListener('mousemove', function(e) {
	var cursorPos = getCursorPos(e);
	var ctx2 = canvas5_2.getContext("2d");

	ctx2.clearRect(0,0,canvas5_2.width,250);
			canvas5_2.style.cursor = "auto";
	for (i=0, len=monthProportion.length; i < len; i++) {
		if (Math.abs(cursorPos.x+circleX[i]) <= 15 && Math.abs(circleY[i]+250-cursorPos.y) <= 15) {
			canvas5_2.style.cursor = "pointer";
			plotPointDetails(ctx2, circleX[i], circleY[i], i);
		}
	}
}, false);

window.onresize = function() {
	if (canvas.getContext) {
		// resizeLineGraph(false);
	} else {
		// Browser is not supported
	}

}

window.onload = function() {
	var values = [33,47,53];
	var values2 = [11,9,4];
	var ctx = canvas.getContext("2d");
	var ctx2 = canvas2.getContext("2d");
	var ctx2_2 = canvas2_2.getContext("2d");
	var ctx3 = canvas3.getContext("2d");
	var ctx4 = canvas4.getContext("2d");
	var ctx4_2 = canvas4_2.getContext("2d");
	if (canvas.getContext) {
		createDonutChart(ctx, ctx2, ctx2_2, values, "Rides Given");
		createDonutChart(ctx3, ctx4, ctx4_2, values2, "Rides Taken");
		resizeLineGraph(true);
	} else {
		// Browser is not supported
	}
}