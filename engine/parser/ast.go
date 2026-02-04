package parser

type CommandType string

const (
	CmdInsert CommandType = "INSERT"
	CmdSelect CommandType = "SELECT"
)

type Command struct {
	Type  CommandType
	Table string
	Data  string
	// Where ayeuna jadi slice (bisa loba kondisi)
	Where []Condition
}

type Condition struct {
	Field    string
	Operator string
	Value    string
	LogicOp  string // "SARENG", "ATAU", atawa kosong (mun kondisi terakhir)
}