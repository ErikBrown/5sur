var inputform = document.getElementById("upload_input")

function checkFileName() {
	var fileName = inputform.value
	var extension = fileName.split(".")[1].toUpperCase()
	if (extension == "PNG" || extension == "JPEG" || extension == "JPG" || extension == "GIF"){
		return true;
	}
	else {
		alert(fileName.split(".")[1] + " tipo de archivo invalido. imágenes debe ser .png, .jpg, o gif no animado.");
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
		alert("Sólo puedes subir un archivo a la vez")
		inputform.value = "";
		return
	}
	var size = files[0].size;
	if (size > 10485760) { // 1MB
		alert("Archivo demasiado pesado. Tamaño máximo  = 10MB")
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
	/*
	img.onload = function(){
		var ratio = img.width / img.height;
		if (ratio < .8 || ratio > 1.2) {
			alert("Invalid image dimensions. Try to use a more square image");
			inputform.value = "";
			return
		}
	};
	*/
	reader.readAsDataURL(files[0]);
}

inputform.addEventListener("change", handleFiles, false);