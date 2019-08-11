package main

import (
    "fmt"
    "strings"
)

func parseCSS(input string) map[string]string {
    output := make(map[string]string)
    properties := strings.Split(input, ";")
    for _, property := range properties {
        propertymap := strings.Split(property, ":")
        output[propertymap[0]] = propertymap[1]
    }
    return output
}

func colorToHTML(input string) string {
    if strings.HasPrefix(input, "#") {
        // already html
        return input
    }
    if strings.HasPrefix(input, "rgb(") {
        r := 0
        g := 0
        b := 0
        fmt.Sscanf(input, "rgb(%d,%d,%d)", &r, &g, &b)
        return strings.ReplaceAll(fmt.Sprintf("#%2x%2x%2x", r, g, b), " ", "0")
    }
    panic("Unsupported color attribute!")
}