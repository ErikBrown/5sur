var inputform = document.getElementById("upload_input")

function checkFileName() {
	var fileName = inputform.value
	var extension = fileName.split(".")[1].toUpperCase()
	if (extension == "PNG" || extension == "JPEG" || extension == "JPG" || extension == "GIF"){
		return true;
	}
	else {
		alert("File with " + fileName.split(".")[1] + " is invalid. Images have to be png, jpg, or non-animated gif");
		inputform.value = "";
		return false;
	}
	return true;
}

function handleFiles() {

	if (!checkFileName()) {
		return false;
	}

	var files = this.files;
	if (files.length > 1) {
		alert("You can only upload one item at a time")
		inputform.value = "";
		return
	}
	var size = files[0].size;
	if (size > 1048576) { // 1MB
		alert("File size too big. Max size = 1MB")
		inputform.value = "";
		return
	}

	hidden = document.getElementById("upload_hidden");



	var img = new Image();

	var reader = new FileReader();
	reader.onload = (function(aImg) { 
		return function(e) { 
			aImg.src = e.target.result;
			
		};
	})(img);
	img.onload = function(){
		var ratio = img.width / img.height;
		if (ratio < .8 || ratio > 1.2) {
			alert("Invalid image dimesions. Try to use a more square image");
			inputform.value = "";
			return
		}
	};
	reader.readAsDataURL(files[0]);
}

inputform.addEventListener("change", handleFiles, false);
