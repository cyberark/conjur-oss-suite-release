{{- if eq (toLower .CertificationLevel) "trusted" -}}
[![Certification Level](https://img.shields.io/badge/Certification%20Level-Trusted-007BFF)]({{ .URL }})
{{- else if eq (toLower .CertificationLevel) "certified" -}}
[![Certification Level](https://img.shields.io/badge/Certification%20Level-Certified-6C757D)]({{ .URL }})
{{- else if eq (toLower .CertificationLevel) "community" -}}
[![Certification Level](https://img.shields.io/badge/Certification%20Level-Community-28A745)]({{ .URL }})
{{- else if eq (toLower .CertificationLevel) "unknown" -}}
[![Certification Level](https://img.shields.io/badge/Certification%20Level-Unknown-DC3545)]({{ .URL }})
{{- end -}}
