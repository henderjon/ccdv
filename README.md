# CCDV

CCDV stands for "Control Character Delimited Value". The format is
essentially meant to be a primitive version of CSVs. Only, instead of
all the random rules around commas, quotes, escapes, etc. I've chosen to
implement it using control characters from the ASCII definition. Namely:

```golang
const (
    // Data Link Escape (DLE)
    Comment = rune('\020')
    // Group Separator (GS)
    GroupSep = rune('\035')
    // Record Separator (RS)
    RecordSep = rune('\036')
    // Unit Separator (US)
    UnitSep = rune('\037')
)
```


The original idea was hatched from the AWK manual in the discussion of
array keys (Section 2.3; page 53). In AWK arrays can't be
nested--you're only allowed one level of depth. However, you can create
array keys that are one string but use the `SUBSEP` variable to
emulate multi-dimensional arrays.

```
The built-in variable SUBSEP contains the value of the subscript-component separator; its default value is not a comma but \034, a value unlikely to appear in normal text.
```

As I dug into control characters, I found [this
article](https://www.lammertbies.nl/comm/info/ascii-characters.html) and noticed
that these characters were designed for this type of thing. Given that
these control characters are unlikely to appear in the values being
stored, this spec could simply disallow them--no more escapes and no
more quotes. The tradeoff, however, is human readability. These
characters don't render as neatly (or at all) when viewing the file in
plain text. This format is quite useful in it's simplicity, but probably
not very useful in the day-to-day. It's a fun concept though. Also, yes,
this was copy-n-pasted from `encoding/csv` in the stdlib. To read the
bytes on the cli, try piping through `tr "\036\037" "\n,"`.
