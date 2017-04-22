package lint

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/ValeLint/vale/core"
	"github.com/russross/blackfriday"
	"golang.org/x/net/html"
	"matloob.io/regexp"
)

// reStructuredText configuration.
//
// reCodeBlock is used to convert Sphinx-style code directives to the regular
// `::` for rst2html.
var reCodeBlock = regexp.MustCompile(`.. (?:raw|code(?:-block)?):: (\w+)`)
var rstArgs = []string{
	"--quiet",
	"--halt=5",
	"--report=5",
	"--link-stylesheet",
	"--no-file-insertion",
	"--no-toc-backlinks",
	"--no-footnote-backlinks",
	"--no-section-numbering",
}

// AsciiDoc configuration.
var adocArgs = []string{
	"-s",
	"--quiet",
	"--safe-mode",
	"secure",
	"-",
}

// Blackfriday configuration.
var commonHTMLFlags = 0 | blackfriday.HTML_USE_XHTML
var commonExtensions = 0 |
	blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
	blackfriday.EXTENSION_TABLES |
	blackfriday.EXTENSION_FENCED_CODE
var renderer = blackfriday.HtmlRenderer(commonHTMLFlags, "", "")
var options = blackfriday.Options{Extensions: commonExtensions}

// HTML configuration.
var heading = regexp.MustCompile(`^h\d$`)

// skipTags are tags that we don't want to lint.
var skipTags = []string{"script", "style", "pre", "code", "tt", "figure"}
var skipClasses = []string{}

func (l Linter) lintHTMLTokens(f *core.File, ctx string, fsrc []byte, offset int) {
	var txt, attr string
	var tokt html.TokenType
	var tok html.Token
	var inBlock, skip, isHeading bool

	lines := len(f.Lines) + offset
	tokens := html.NewTokenizer(bytes.NewReader(fsrc))
	for {
		tokt = tokens.Next()
		tok = tokens.Token()
		txt = core.PrepText(html.UnescapeString(strings.TrimSpace(tok.Data)))
		skip = core.StringInSlice(txt, skipTags) || core.StringInSlice(attr, skipClasses)
		if tokt == html.ErrorToken {
			break
		} else if tokt == html.StartTagToken && skip {
			inBlock = true
		} else if skip && inBlock {
			inBlock = false
		} else if tokt == html.StartTagToken && heading.MatchString(txt) {
			isHeading = true
		} else if tokt == html.CommentToken {
			f.UpdateComments(txt)
		} else if tokt == html.EndTagToken && isHeading {
			isHeading = false
		} else if tokt == html.TextToken && isHeading && !inBlock && txt != "" {
			l.lintText(f, NewBlock(ctx, txt, "text.heading"+f.RealExt), lines, 0)
		} else if tokt == html.TextToken && !inBlock && !skip && txt != "" {
			l.lintProse(f, ctx, txt, lines, 0)
		}
		attr = getAttribute(tok, "class")
		ctx = clearElements(ctx, tok)
		ctx = updateCtx(ctx, txt, tokt)
	}
}

func getAttribute(tok html.Token, key string) string {
	for _, attr := range tok.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func clearElements(ctx string, tok html.Token) string {
	if tok.Data == "img" || tok.Data == "a" {
		for _, a := range tok.Attr {
			if a.Key == "alt" || a.Key == "href" {
				ctx = updateCtx(ctx, a.Val, html.TextToken)
			}
		}
	}
	return ctx
}

func updateCtx(ctx string, txt string, tokt html.TokenType) string {
	var found bool
	if (tokt == html.TextToken || tokt == html.CommentToken) && txt != "" {
		for _, s := range strings.Split(txt, "\n") {
			ctx, found = core.Substitute(ctx, s)
			if !found {
				for _, w := range strings.Fields(s) {
					ctx, _ = core.Substitute(ctx, w)
				}
			}
		}
	}
	return ctx
}

func (l Linter) lintHTML(f *core.File) {
	l.lintHTMLTokens(f, f.Content, []byte(f.Content), 0)
}

func (l Linter) lintMarkdown(f *core.File) {
	html := blackfriday.MarkdownOptions([]byte(f.Content), renderer, options)
	l.lintHTMLTokens(f, f.Content, html, 0)
}

func (l Linter) lintRST(f *core.File, python string, rst2html string) {
	var out bytes.Buffer
	cmd := exec.Command(python, append([]string{rst2html}, rstArgs...)...)
	cmd.Stdin = strings.NewReader(reCodeBlock.ReplaceAllString(f.Content, "::"))
	cmd.Stdout = &out
	if core.CheckError(cmd.Run()) {
		l.lintHTMLTokens(f, f.Content, out.Bytes(), 0)
	}
}

func (l Linter) lintADoc(f *core.File, asciidoctor string) {
	var out bytes.Buffer
	cmd := exec.Command(asciidoctor, adocArgs...)
	cmd.Stdin = strings.NewReader(f.Content)
	cmd.Stdout = &out
	if core.CheckError(cmd.Run()) {
		l.lintHTMLTokens(f, f.Content, out.Bytes(), 0)
	}
}
