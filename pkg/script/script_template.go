package script

const scriptTemplate = `#!/usr/bin/bash

usage() {
	echo "netnscli usage:"
	echo ""
	echo "$0 [ACTION]"
	echo ""
	echo "ACTIONS"
	echo "		apply		Apply the testbed"
	echo ""
	echo "		delete		delete the testbed"
	echo ""
}

die() {
	echo "Unrecoverable error: $1"
	exit 1
}

apply() {
	# add namespaces
	{{range .Namespaces}}ip netns add {{.Name}}
	{{end}}
	# Create veth pairs
	{{- range .VethPairs}}
	ip link add {{.P1.Name}} type veth peer name {{.P2.Name}}

	{{- if isNotDefaultNamespace .P1.Namespace}}
	ip link set {{.P1.Name}} netns {{.P1.Namespace}}{{end}}
	{{- if isNotDefaultNamespace .P2.Namespace}}
	ip link set {{.P2.Name}} netns {{.P2.Namespace}}{{end}}

	{{- if .P1.Address}}
	{{- if isNotDefaultNamespace .P1.Namespace}}
	ip netns exec {{.P1.Namespace}} ip addr add {{.P1.Address}} dev {{.P1.Name}}
	{{- else}}
	ip addr add {{.P1.Address}} dev {{.P1.Name}}
	{{- end}}
	{{- end}}
	{{- if .P2.Address}}
	{{- if isNotDefaultNamespace .P2.Namespace}}
	ip netns exec {{.P2.Namespace}} ip addr add {{.P2.Address}} dev {{.P2.Name}}
	{{- else}}
	ip addr add {{.P2.Address}} dev {{.P2.Name}}
	{{- end}}
	{{- end}}

	{{- if isNotDefaultNamespace .P1.Namespace}}
	ip netns exec {{.P1.Namespace}} ip link set {{.P1.Name}} up
	{{- else}}
	ip link set {{.P1.Name}} up
	{{- end}}
	{{- if isNotDefaultNamespace .P2.Namespace}}
	ip netns exec {{.P2.Namespace}} ip link set {{.P2.Name}} up
	{{- else}}
	ip link set {{.P2.Name}} up
	{{- end}}
	{{end}}
	# Create and configure bridges
	{{- range $b := .Bridges}}
	ip link add name {{$b.Name}} type bridge
	ip link set {{$b.Name}} up
	{{- range $i := .Interfaces}}
	ip link set {{$i}} master {{$b.Name}}
	{{- end}}
	{{end}}
}

delete() {
	{{- range $b := .Bridges}}
	{{- range $i := .Interfaces}}
	ip link set {{$i}} nomaster
	{{- end}}
	{{- end}}
	{{range .Namespaces}}
	ip netns del {{.Name}}
	{{- end}}
	{{range .Bridges}}
	ip link del {{.Name}}
	{{- end}}
}

if [ -z $1 ]; then
	die "You must specify an action between apply and delete"
fi

if [ $1 == "apply" ]; then
	apply
elif [ $1 == "delete" ]; then
	delete
else
	die "$1 is not an existing netnscli action"
fi
`
