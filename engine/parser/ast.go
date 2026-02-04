package parser

type CommandType string

const (
	CmdCreate CommandType = "CREATE"
	CmdInsert CommandType = "INSERT"
	CmdSelect CommandType = "SELECT"
	CmdUpdate CommandType = "UPDATE"
	CmdDelete CommandType = "DELETE"
)

type JoinClause struct {
    Type      string // "INNER", "LEFT", "RIGHT"
    Table     string // Tabel nu rek digabung (misal: divisi)
    Condition Condition // Kondisi ON (misal: pegawai.divisi_id = divisi.id)
}

type Command struct {
	Type    CommandType
	Table   string
	Fields  []string
	Data    string    
	Updates map[string]string 
	Where   []Condition
	Condition []Condition
	Joins 	[]JoinClause
	OrderBy   string 
	OrderDesc bool   
	Limit     int   
	Offset    int    
}

type Condition struct {
	Field    string
	Operator string
	Value    string
	LogicOp  string
}