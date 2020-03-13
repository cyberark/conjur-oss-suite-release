{{- if eq (toLower .CertificationLevel) "trusted" -}}
[![Certification Level](https://img.shields.io/badge/Certification%20Level-Trusted-Blue)]({{ .URL }})
{{- else if eq (toLower .CertificationLevel) "certified" -}}
[![Certification Level](https://img.shields.io/badge/Certification%20Level-Certified-Green)]({{ .URL }})
{{- else if eq (toLower .CertificationLevel) "community" -}}
[![Certification Level](https://img.shields.io/badge/Certification%20Level-Community-Yellow)]({{ .URL }})
{{- else -}}
[![Certification Level](https://img.shields.io/badge/Certification%20Level-Unknown-Red)]({{ .URL }})
{{- end -}}
