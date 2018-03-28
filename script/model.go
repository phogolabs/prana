package script

var (
	format = "20060102150405"
)

type FileGenerator interface {
	Create(container, command string) (string, error)
}
