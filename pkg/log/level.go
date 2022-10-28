package log

// Level of severity.
type Level int

// Verbose is a boolean type that implements Info, Infov (like Printf) etc.
type Verbose bool

// common log level.
const (
	_fatalLevel Level = iota
	_errorLevel
	_warnLevel
	_infoLevel
	_debugLevel
)

var levelNames = [...]string{
	_fatalLevel: RedBold + "[FATAL]" + Reset,
	_errorLevel: Red + "[ERROR]" + Reset,
	_warnLevel:  Magenta +"[WARN]" + Reset,
	_infoLevel:  Green + "[INFO]" + Reset,
	_debugLevel: Yellow + "[DEBUG]" + Reset,
}

// String implementation.
func (l Level) String() string {
	return levelNames[l]
}
