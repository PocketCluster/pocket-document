// http://stackoverflow.com/questions/34112382/golang-issues-replacing-newlines-in-a-string-from-a-text-file

func regexpNewLine() {
    re = regexp.MustCompile(`\r?\n`)
    input = re.ReplaceAllString(input, " ")    
}

// http://stackoverflow.com/questions/38305668/golang-removing-all-unicode-newline-characters-from-a-string

func filterNewLines(s string) string {
    return strings.Map(func(r rune) rune {
        switch r {
        case 0x000A, 0x000B, 0x000C, 0x000D, 0x0085, 0x2028, 0x2029:
            return -1
        default:
            return r
        }
    }, s)
}