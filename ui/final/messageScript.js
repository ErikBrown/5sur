var textArea = document.getElementById('message_text');

textArea.addEventListener("keydown", function() {
	textLength()
}, false);

function textLength() {
	console.log(textArea.value.length);
}

textArea.addEventListener("blur", function() {
	textLength()
}, false);