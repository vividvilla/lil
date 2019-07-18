<html prefix="og: http://ogp.me/ns#">
<head>
	<title>{{ if .Title }}{{ .Title }}{{ else }} Redirect {{ end }}</title>
	{{- if .OGTags }}
	{{- range .OGTags }}
	<meta property="{{ .Property }}" content="{{ .Content }}" />
	{{- end}}
	{{- end}}
	<script >
		window.location.replace("{{ .URL }}");
	</script>
</head>
</html>
