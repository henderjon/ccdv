package ccdv

const (
	// Comment is the ascii Data Link Escape (DLE) character and begins a commented record
	Comment = rune('\020') // byte(16) || "\x10" ...
	// GroupSep is the ascii Group Separator (GS) character and delineates groups of records (think DB tables)
	GroupSep = rune('\035') // "\035" byte(29) || "\x1d" ...
	// RecordSep is the ascii Record Separator (RS) character and delineates records (think DB rows in a table)
	RecordSep = rune('\036') // "\036" byte(30) || "\x1e" ...
	// UnitSep is the ascii Unit Separator (US) character and delineates fields in a record (think DB fields in a row)
	UnitSep = rune('\037') // "\037" byte(31) || "\x1f" ...
	// lammertbies.nl/comm/info/ascii-characters.html
)
