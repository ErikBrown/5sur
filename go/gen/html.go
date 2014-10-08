package gen

func HeaderHtml(title string) string {
return `<!doctype html>
<html>
	<head>
		<title>` + title + `</title>
		<link href="https://fonts.googleapis.com/css?family=Montserrat:400,700|Open+Sans:400,400italic,600,300,700,800|Bitter:400,400italic,700" rel="stylesheet" type="text/css">
		<link rel="stylesheet" type="text/css" href="https://192.241.219.35/style.css" />
	</head>
<body>
<div id="header">
	<h1><a href="https://192.241.219.35">RideChile</a></h1>
	<ul id="account_nav">
		<li>UserName</li>
		<li>MsgIcon</li>
		<li>Logout</li>
	</ul>
</div>`
}

func FooterHtml() string{
return `</body>
</html>`
}

func Error404() string{
	return `
	<!doctype html>

	<html>

	<head>

	<meta charset='utf-8'>

	<link rel="stylesheet" type="text/css" href="https://192.241.219.35/404_style.css" />
	<link href='https://fonts.googleapis.com/css?family=Exo:700' rel='stylesheet' type='text/css'>
	<link href='https://fonts.googleapis.com/css?family=Open+Sans' rel='stylesheet' type='text/css'>

	<title>RideShare App</title>

	</head>

	<body>
	<div id="center">
		<h1>404 Error</h1>
		<span><a href="https://192.241.219.35/">Return to homepage</a></span>
	</div>
	</body>

	</html>
	`
}