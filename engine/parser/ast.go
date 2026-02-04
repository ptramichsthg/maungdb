package parser

type CommandType string

const (
	CmdInsert CommandType = "INSERT"
	CmdSelect CommandType = "SELECT"
	CmdUpdate CommandType = "UPDATE" // Anyar: OMEAN
	CmdDelete CommandType = "DELETE" // Anyar: MICEUN
)

type Command struct {
	Type    CommandType
	Table   string
	Data    string      // Pikeun INSERT
	Updates map[string]string // Pikeun UPDATE (col=val) -> Anyar
	Where   []Condition
}

type Condition struct {
	Field    string
	Operator string
	Value    string
	LogicOp  string
}