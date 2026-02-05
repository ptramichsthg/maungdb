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
    Type      string
    Table     string 
    Condition Condition 
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