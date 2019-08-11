package main

import (
    "fmt"
    "io/ioutil"
    "encoding/json"
    "encoding/base64"
    "net/url"
    "time"
    "github.com/JoshVarga/svgparser"
    "os"
    "strings"
    "strconv"
    "github.com/rustyoz/genericlexer"
)

const sessionId = "NOPE"

const pointSlopeFormat = "-y-[y1]=\\frac{[y2]-[y1]}{[x2]-[x1]}\\left(x-[x1]\\right)\\left\\{[left]<x<[right]\\right\\}\\left\\{[bottom]<y<[top]\\right\\}"
//const verticalLineFormat = "x=[x1]\\left\\{-[bottom]<y<[top]\\right\\}"
//const horizontalLineFormat = "y=-[y1]\\left\\{[left]<x<[right]\\right\\}"
const verticalLineFormat = "x=v\\left(y,[x1],-[bottom],[top]\\right)"
const horizontalLineFormat = "y=h\\left(x,-[y1],[left],[right]\\right)"
const cubicBezierFormat = "\\left(B_x\\left(t,[x1],[x2],[x3],[x4]\\right),B_y\\left(t,[y1],[y2],[y3],[y4]\\right)\\right)"
const ellipseFormat = "\\frac{\\left(x-[cx]\\right)^2}{[rx]^2}+\\frac{\\left(y+[cy]\\right)^2}{[ry]^2}=1"

type commonFunction int
const (
    verticalLine commonFunction = iota
    horizontalLine commonFunction = iota
    cubicBezierX commonFunction = iota
    cubicBezierY commonFunction = iota
    ellipse commonFunction = iota
)

var commonFunctions = [...]string {
    "v\\left(y,p,b,t\\right)=p\\left\\{b<y<t\\right\\}",
    "h\\left(x,p,l,r\\right)=p\\left\\{l<x<r\\right\\}",
    "B_x\\left(t,c_1,c_2,c_3,c_4\\right)=\\left(1-t\\right)^3c_1+3t\\left(1-t\\right)^2c_2+3t^2\\left(1-t\\right)c_3+t^3c_4\\ \\left\\{0<t\\le1\\right\\}",
    "B_y\\left(t,v_1,v_2,v_3,v_4\\right)=\\left(1-t\\right)^3v_1+3t\\left(1-t\\right)^2v_2+3t^2\\left(1-t\\right)v_3+t^3v_4",
}

// from github.com/rustyoz/svg
type Tuple [2]float64

func parseTuple(l *genericLexer.Lexer) (Tuple, error) {
    t := Tuple{}

    l.ConsumeWhiteSpace()

    ni := l.NextItem()
    if ni.Type == genericLexer.ItemNumber {
        n, ok := strconv.ParseFloat(ni.Value, 64)
        if ok != nil {
            return t, fmt.Errorf("Error parsing number %s", ok)
        }
        t[0] = n
    } else {
        return t, fmt.Errorf("Error parsing Tuple expected Number got: %s", ni.Value)
    }

    if l.PeekItem().Type == genericLexer.ItemWSP || l.PeekItem().Type == genericLexer.ItemComma {
        l.NextItem()
    }
    ni = l.NextItem()
    if ni.Type == genericLexer.ItemNumber {
        n, ok := strconv.ParseFloat(ni.Value, 64)
        if ok != nil {
            return t, fmt.Errorf("Error passing Number %s", ok)
        }
        t[1] = n
    } else {
        return t, fmt.Errorf("Error passing Tuple expected Number got: %v", ni)
    }

    return t, nil
}

func main() {
    // not the best solution, but i'm lazy
    graph := &desmosGraph{
        Version: 7,
    }
    graph.Graph.Viewport.Xmin = -10
    graph.Graph.Viewport.Ymin = -10
    graph.Graph.Viewport.Xmax = 10
    graph.Graph.Viewport.Ymax = 10
    // append common functions first
    curExpressionId := 0
    for _, function := range commonFunctions {
        graph.Expressions.List = append(graph.Expressions.List, desmosExpression{
            Type: "expression",
            Id: strconv.Itoa(curExpressionId),
            Color: "#000000",
            Latex: function,
        })
        curExpressionId++
    }

    // parse svg
    file, err := os.Open("gopher.svg")
    if err != nil {
        panic(err)
    }

    svg, err := svgparser.Parse(file, false)
    if err != nil {
        panic(err)
    }

    for _, element := range svg.Children {
        var expression = &desmosExpression{
            Type: "expression",
        }
        switch element.Name {
        case "line":
            x1 := element.Attributes["x1"]
            x2 := element.Attributes["x2"]
            y1 := element.Attributes["y1"]
            y2 := element.Attributes["y2"]
            width, _ := strconv.ParseFloat(svg.Attributes["width"], 64)
            height, _ := strconv.ParseFloat(svg.Attributes["height"], 64)
            x1Int, _ := strconv.ParseFloat(element.Attributes["x1"], 64)
            x2Int, _ := strconv.ParseFloat(element.Attributes["x2"], 64)
            y1Int, _ := strconv.ParseFloat(element.Attributes["y1"], 64)
            y2Int, _ := strconv.ParseFloat(element.Attributes["y2"], 64)

            style := parseCSS(element.Attributes["style"])
            expression.Color = colorToHTML(style["stroke"])

            if (x2 == x1) {
                // vertical line
                expression.Latex = strings.NewReplacer(
                    "[x1]", x1,
                    "[top]", float64ToString(min(y1Int, y2Int)),
                    "[bottom]", float64ToString(max(y1Int, y2Int)),
                ).Replace(verticalLineFormat)
                break
            }

            if (y2 == y1) {
                // horizontal line
                expression.Latex = strings.NewReplacer(
                    "[y1]", y1,
                    "[left]", float64ToString(min(x1Int, x2Int)),
                    "[right]", float64ToString(max(x1Int, x2Int)),
                ).Replace(horizontalLineFormat)
                break
            }

            // point slope formula
            // -y - y1 = ((y2 - y1) / (x2 - x1))(x - x1)
            expression.Latex = strings.NewReplacer(
                "[x1]", x1,
                "[x2]", x2,
                "[y1]", y1,
                "[y2]", y2,
                "[top]", "0",
                "[bottom]", float64ToString(-(height)),
                "[left]", "0",
                "[right]", float64ToString(width),
            ).Replace(pointSlopeFormat)
        case "path":
            // VERY unfinished!!!! this will only parse gopher.svg!!!!
            expression.Color = "#000000"
            curPos := Tuple{0, 0}
            lastPos := Tuple{0, 0}
            l, _ := genericLexer.Lex(strconv.Itoa(curExpressionId), element.Attributes["d"])
            lex := *l
            func() {
                for {
                    i := lex.NextItem()
                    switch {
                    case i.Type == genericLexer.ItemError:
                        fallthrough
                    case i.Type == genericLexer.ItemEOS:
                        return
                    case i.Type == genericLexer.ItemLetter:
                        fmt.Println(i.Value)
                        switch i.Value {
                        case "M":
                            // moveto
                            curPos, err := parseTuple(&lex)
                            lastPos = curPos
                            if err != nil {
                                panic(err)
                            }
                        case "C":
                            // cubic bezier curve (absolute)
                            var curveExpression = &desmosExpression{
                                Type: "expression",
                            }
                            curveExpression.Color = "#000000"
                            var points []Tuple
                            for lex.PeekItem().Type == genericLexer.ItemNumber {
                                t, _ := parseTuple(&lex)
                                points = append(points, t)
                                lex.ConsumeWhiteSpace()
                                lex.ConsumeComma()
                            }
                            if (len(points) != 3) {
                                fmt.Println(len(points))
                                panic("unexpected point count")
                            }
                            curveExpression.Latex = strings.NewReplacer(
                                "[x1]", float64ToString(curPos[0]),
                                "[y1]", float64ToString(curPos[1] * -1),
                                "[x2]", float64ToString(points[0][0]),
                                "[y2]", float64ToString(points[0][1] * -1),
                                "[x3]", float64ToString(points[1][0]),
                                "[y3]", float64ToString(points[1][1] * -1),
                                "[x4]", float64ToString(points[2][0]),
                                "[y4]", float64ToString(points[2][1] * -1),
                            ).Replace(cubicBezierFormat)
                            fmt.Println(curveExpression.Latex)
                            curPos = Tuple(points[2])
                            curveExpression.Id = strconv.Itoa(curExpressionId)
                            curExpressionId++
                            graph.Expressions.List = append(graph.Expressions.List, *curveExpression)
                        case "c":
                            // cubic bezier curve (relative)
                            var curveExpression = &desmosExpression{
                                Type: "expression",
                            }
                            curveExpression.Color = "#000000"
                            var points []Tuple
                            for lex.PeekItem().Type == genericLexer.ItemNumber {
                                t, _ := parseTuple(&lex)
                                points = append(points, t)
                                lex.ConsumeWhiteSpace()
                                lex.ConsumeComma()
                            }
                            if (len(points) != 3) {
                                fmt.Println(len(points))
                                panic("unexpected point count")
                            }
                            curveExpression.Latex = strings.NewReplacer(
                                "[x1]", float64ToString(curPos[0]),
                                "[y1]", float64ToString(curPos[1] * -1),
                                "[x2]", float64ToString(points[0][0]),
                                "[y2]", float64ToString(points[0][1] * -1),
                                "[x3]", float64ToString(points[1][0]),
                                "[y3]", float64ToString(points[1][1] * -1),
                                "[x4]", float64ToString(points[2][0]),
                                "[y4]", float64ToString(points[2][1] * -1),
                            ).Replace(cubicBezierFormat)
                            fmt.Println(curveExpression.Latex)
                            curPos = Tuple(points[2])
                            curveExpression.Id = strconv.Itoa(curExpressionId)
                            curExpressionId++
                            graph.Expressions.List = append(graph.Expressions.List, *curveExpression)
                        case "Z":
                            fallthrough
                        case "z":
                            fmt.Println(curPos)
                            fmt.Println(lastPos)
                            x1 := curPos[0]
                            x2 := lastPos[0]
                            y1 := curPos[1]
                            y2 := lastPos[1]

                            expression.Color = "#000000"

                            if (x2 == x1) {
                                // vertical line
                                expression.Latex = strings.NewReplacer(
                                    "[x1]", float64ToString(x1),
                                    "[top]", float64ToString(min(y1, y2)),
                                    "[bottom]", float64ToString(max(y1, y2)),
                                ).Replace(verticalLineFormat)
                                break
                            }

                            if (y2 == y1) {
                                // horizontal line
                                expression.Latex = strings.NewReplacer(
                                    "[y1]", float64ToString(y1),
                                    "[left]", float64ToString(min(x1, x2)),
                                    "[right]", float64ToString(max(x1, x2)),
                                ).Replace(horizontalLineFormat)
                                break
                            }

                            // point slope formula
                            // -y - y1 = ((y2 - y1) / (x2 - x1))(x - x1)
                            expression.Latex = strings.NewReplacer(
                                "[x1]", float64ToString(x1),
                                "[x2]", float64ToString(x2),
                                "[y1]", float64ToString(y1),
                                "[y2]", float64ToString(y2),
                                "[top]", float64ToString(-(min(y1, y2))),
                                "[bottom]", float64ToString(-(max(y1, y2))),
                                "[left]", float64ToString(min(x1, x2)),
                                "[right]", float64ToString(max(x1, x2)),
                            ).Replace(pointSlopeFormat)
                        case "L":
                        default:
                            panic(i.Value)
                        }
                    default:
                    }
                }
            }()
        case "ellipse":
            expression.Color = "#000000"
            expression.Latex = strings.NewReplacer(
                "[cx]", element.Attributes["cx"],
                "[rx]", element.Attributes["rx"],
                "[cy]", element.Attributes["cy"],
                "[ry]", element.Attributes["ry"],
            ).Replace(ellipseFormat)
        default:
            panic(element.Name)
        }
        expression.Id = strconv.Itoa(curExpressionId)
        curExpressionId++
        graph.Expressions.List = append(graph.Expressions.List, *expression)
    }

    // create graph json
    graphJson, err := json.Marshal(graph)
    if err != nil {
        panic(err)
    }

    // read the thumbnail file
    data, err := ioutil.ReadFile("default.png")
    if err != nil {
        panic(err)
    }
    pngBase64 := base64.StdEncoding.EncodeToString(data)

    // format graph upload
    postData := fmt.Sprintf("thumb_data=%s%s&graph_hash=%s&version=%s&my_graphs=true&is_update=false&title=%s&calc_state=%s",
        url.QueryEscape("data:image/png;base64,"),
        url.QueryEscape(pngBase64),
        desmosRandomHash(),
        "h3",
        fmt.Sprint(time.Now().Unix()),
        url.QueryEscape(string(graphJson)),
    )

    // upload!
    fmt.Println(desmosPost("/api/v1/calculator/save", postData))
    //var _ = postData // only to make go shut up when above is commented out
    fmt.Println(string(graphJson))
}