package main

import "text/template"

// export enum PermResult {
//   UNSET = 0,
//   ALLOWED = 1,
//   DENIED = 2
// }

var tmplTypeScript = `
import type { PermResult } from "./types-perm";


export interface {{.MapName}} {
  {{- range .Perms}}
  {{.Key}}: PermResult; // {{.Desc}}
  {{- end}}
}
`

var tmplTs = template.Must(template.New("perm").Parse(tmplTypeScript))
