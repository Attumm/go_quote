package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var VoiceFormats = map[string][]string{
	"aiff": {"Albert", "Alice", "Amira", "Anna", "Bad", "Bells", "Boing", "Carmit", "Cellos", "Damayanti", "Daniel", "Wobble", "Eddy", "Ellen", "Flo", "Fred", "Good", "Grandma", "Grandpa", "Jester", "Jacques", "Joana", "Junior", "Kanya", "Karen", "Kyoko", "Laura", "Lekha", "Lesya", "Luciana", "Majed", "Meijia", "Melina", "Moira", "Organ", "Superstar", "Ralph", "Reed", "Rishi", "Rocko", "Samantha", "Sara", "Shelley", "Sinji", "Tessa", "Thomas", "Trinoids", "Whisper", "Xander", "Yuna", "Zarvox", "Zosia"},
	"mp3":  {"Alice", "Amira", "Anna", "Carmit", "Damayanti", "Daniel", "Eddy", "Ellen", "Flo", "Good", "Grandma", "Grandpa", "Jacques", "Joana", "Kanya", "Karen", "Kyoko", "Laura", "Lekha", "Lesya", "Luciana", "Majed", "Meijia", "Melina", "Moira", "Reed", "Rishi", "Rocko", "Samantha", "Sara", "Shelley", "Sinji", "Tessa", "Thomas", "Xander", "Yuna", "Zarvox", "Zosia", "Zuzana"},
	"ogg":  {"Albert", "Alice", "Amira", "Anna", "Bad", "Bells", "Boing", "Carmit", "Cellos", "Damayanti", "Daniel", "Wobble", "Eddy", "Ellen", "Flo", "Fred", "Good", "Grandma", "Grandpa", "Jester", "Jacques", "Joana", "Junior", "Kanya", "Karen", "Kyoko", "Laura", "Lekha", "Lesya", "Luciana", "Majed", "Meijia", "Melina", "Moira", "Organ", "Superstar", "Ralph", "Reed", "Rishi", "Rocko", "Samantha", "Sara", "Shelley", "Sinji", "Tessa", "Thomas", "Trinoids", "Whisper", "Xander", "Yuna", "Zarvox", "Zosia"},
	"wav":  {"Albert", "Alice", "Amira", "Anna", "Bad", "Bells", "Boing", "Carmit", "Cellos", "Damayanti", "Daniel", "Wobble", "Eddy", "Ellen", "Flo", "Fred", "Good", "Grandma", "Grandpa", "Jester", "Jacques", "Joana", "Junior", "Kanya", "Karen", "Kyoko", "Laura", "Lekha", "Lesya", "Luciana", "Majed", "Meijia", "Melina", "Moira", "Organ", "Superstar", "Ralph", "Reed", "Rishi", "Rocko", "Samantha", "Sara", "Shelley", "Sinji", "Tessa", "Thomas", "Trinoids", "Whisper", "Xander", "Yuna", "Zarvox", "Zosia"},
}

func getRandomVoice(format string) string {
	voices, exists := VoiceFormats[format]
	if !exists || len(voices) == 0 {
		return "Alice"
	}
	return voices[rand.Intn(len(voices))]
}

func serveAudioQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo, format string) {
	// security check, if format is known
	w.Header().Set("Content-Type", OutputFormats[format])
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=speech.%s", format))

	internalFilename := fmt.Sprintf("speech_%d.%s", time.Now().UnixNano(), format)
	text := fmt.Sprintf("%s. By %s", q.Text, q.Author)
	tempFile := filepath.Join(os.TempDir(), internalFilename)

	var cmd *exec.Cmd

	// good ones
	goodVoices := []string{
		"Whisper",
		"Majed",
		"Superstar",
		"Ralph",
		"Kyoko",
	}
	goodVoices = append(goodVoices, q.Text)

	singingVoices := []string{
		"Organ",
		"Bad",
		"Good",
	}
	singingVoices = append(singingVoices, q.Text)
	scary := []string{
		"Trinoids",
	}
	scary = append(scary, q.Text)
	if api.Runtime == "darwin" {
		voice := getRandomVoice(format)
		fmt.Println("Used voice:", voice)
		args := []string{"-o", tempFile, "-v", voice}
		if format == "wav" {
			args = append(args, "--data-format=LEF32@32000")
		}
		args = append(args, text)
		cmd = exec.Command("say", args...)
	} else {
		tempFile = filepath.Join("/usr/share/pico/lang", internalFilename)
		cmd = exec.Command("pico2wave", "-w", tempFile, text)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error generating speech: %v, output: %s", err, output)
		http.Error(w, "Error generating speech", http.StatusInternalServerError)
		return
	}

	audioFile, err := os.Open(tempFile)
	if err != nil {
		log.Printf("Error opening audio file: %v", err)
		http.Error(w, "Error reading audio file", http.StatusInternalServerError)
		return
	}
	defer audioFile.Close()
	defer os.Remove(tempFile)

	fileInfo, err := audioFile.Stat()
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		http.Error(w, "Error reading audio file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	_, err = io.Copy(w, audioFile)
	if err != nil {
		fmt.Printf("Error sending audio: %v", err)
		http.Error(w, "Error sending audio", http.StatusInternalServerError)
		return
	}

}

func serveWavQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	serveAudioQuote(w, q, api, requestData, "wav")
}
func serveAiffQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	if api.Runtime != "darwin" {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		return
	}
	serveAudioQuote(w, q, api, requestData, "aiff")
}
func serveMP3Quote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	if api.Runtime != "darwin" {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		return
	}
	serveAudioQuote(w, q, api, requestData, "mp3")
}
func serveOggQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	if api.Runtime != "darwin" {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		return
	}
	serveAudioQuote(w, q, api, requestData, "ogg")
}

func quoteToSVG(q ResponseQuote) string {
	maxChars := 200
	maxTags := 5
	svgTemplate := `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 500 %d">
        <defs>
            <style>
                .quote { font-family: 'Merriweather', serif; font-size: 24px; fill: #333; }
                .author { font-family: 'Open Sans', sans-serif; font-size: 18px; fill: #666; }
                .tag { font-family: 'Open Sans', sans-serif; font-size: 14px; fill: #333; }
                .metadata { font-family: 'Open Sans', sans-serif; font-size: 12px; fill: #999; }
            </style>
        </defs>
        <rect width="100%%" height="100%%" fill="#f8f8f8"/>
        <text class="quote">%s</text>
        <text x="25" y="%d" class="author">%s</text>
        %s
        <text x="25" y="%d" class="metadata">ID: %d</text>
    </svg>`

	processedText := processQuoteText(maxChars, q.Text)

	words := strings.Fields(processedText)
	var lines []string
	currentLine := ""
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= 40 { // +1 for space
			currentLine += word + " "
		} else {
			lines = append(lines, strings.TrimSpace(currentLine))
			currentLine = word + " "
		}
	}
	if currentLine != "" {
		lines = append(lines, strings.TrimSpace(currentLine))
	}

	quoteText := ""
	lineHeight := 30 // Reduced from previous value
	for i, l := range lines {
		quoteText += fmt.Sprintf(`<tspan x="25" dy="%d">%s</tspan>`, lineHeight, l)
		if i == 0 {
			lineHeight = 28 // Slightly reduce space between subsequent lines
		}
	}

	authorY := 40 + len(lines)*lineHeight + 20

	processedTags := processTags(maxTags, q.Tags)

	tagsSVG := ""
	for i, tag := range processedTags {
		tagsSVG += fmt.Sprintf(`
            <rect x="%d" y="%d" width="80" height="24" rx="12" fill="#e0e0e0"/>
            <text x="%d" y="%d" class="tag" text-anchor="middle">%s</text>`,
			25+i*90, authorY+30, 65+i*90, authorY+46, tag)
	}

	svgHeight := authorY + 90 + len(processedTags)*30

	metadataY := svgHeight - 20

	authorName := ""
	if q.Author != "" {
		authorName = "— " + q.Author
	}

	return fmt.Sprintf(svgTemplate, svgHeight, quoteText, authorY, authorName, tagsSVG, metadataY, q.ID)
}

func processQuoteText(maxChar int, text string) string {
	// Remove all quotation marks
	text = strings.ReplaceAll(text, "\"", "")
	text = strings.ReplaceAll(text, "'", "")

	// Truncate to 150 characters and add ellipsis if needed
	if len(text) > maxChar {
		text = text[:maxChar-3] + "..."
	}

	return text
}

func processTags(maxTags int, tags []string) []string {
	if len(tags) > maxTags {
		tags = tags[:maxTags]
	}

	for i, tag := range tags {
		// Remove hyphens and convert to lowercase
		tag = strings.ReplaceAll(tag, "-", " ")
		tag = strings.ToLower(tag)
		tags[i] = tag
	}

	return tags
}

func quoteToCSV(quote ResponseQuote) string {
	var sb strings.Builder
	sb.WriteString(`"ID","Text","Author","Tags"`)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`"%d","%s","%s","%s"`,
		quote.ID,
		strings.ReplaceAll(quote.Text, `"`, `""`),
		strings.ReplaceAll(quote.Author, `"`, `""`),
		strings.Join(quote.Tags, "|")))
	return sb.String()
}

func quoteToEmbeddedHTML(quote ResponseQuote) string {
	return fmt.Sprintf(`
<blockquote class="quote-embed" 
    style="font-family: Arial, sans-serif; max-width: 500px; margin: 20px auto; padding: 20px;
           border-left: 5px solid #eee; background-color: #f9f9f9;">
    <p style="font-size: 18px; line-height: 1.4; margin-bottom: 10px;">"%s"</p>
    <footer style="color: #666;">
        — <cite>%s</cite>
    </footer>
    <div style="margin-top: 10px; font-size: 14px; color: #888;">
        ID: %d • %s
    </div>
</blockquote>`, quote.Text, quote.Author, quote.ID, strings.Join(quote.Tags, " • "))
}

type OEmbedResponse struct {
	Type         string `json:"type"`
	Version      string `json:"version"`
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url,omitempty"`
	ProviderName string `json:"provider_name"`
	ProviderURL  string `json:"provider_url"`
	CacheAge     int    `json:"cache_age"`
	HTML         string `json:"html"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

func (oe *OEmbedResponse) Create(quote ResponseQuote, baseURL string) {
	oe.Type = "rich"
	oe.Version = "1.0"
	oe.Title = "Quote by " + quote.Author
	oe.AuthorName = quote.Author
	oe.ProviderName = "Quotes API"
	oe.ProviderURL = baseURL
	oe.CacheAge = 3600
	oe.HTML = quoteToEmbeddedHTML(quote)
	oe.Width = 500
	oe.Height = 200
}

type AtomFeed struct {
	XMLName xml.Name `xml:"feed"`
	XMLNS   string   `xml:"xmlns,attr"`
	Title   string   `xml:"title"`
	Link    struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
	} `xml:"link"`
	Updated string     `xml:"updated"`
	Author  AtomAuthor `xml:"author"`
	ID      string     `xml:"id"`
	Entry   AtomEntry  `xml:"entry"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

type AtomEntry struct {
	Title     string     `xml:"title"`
	Link      AtomLink   `xml:"link"`
	ID        string     `xml:"id"`
	Updated   string     `xml:"updated"`
	Content   string     `xml:"content"`
	Published string     `xml:"published"`
	Author    AtomAuthor `xml:"author"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

func (af *AtomFeed) Create(quote ResponseQuote, baseURL, quoteURL string) {
	now := time.Now().Format(time.RFC3339)

	af.XMLNS = "http://www.w3.org/2005/Atom"
	af.Title = "Random Quote"
	af.Updated = now
	af.Author = AtomAuthor{Name: "Quotes API"}
	af.ID = quoteURL
	af.Link = struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
	}{
		Href: baseURL,
		Rel:  "self",
	}
	af.Entry = AtomEntry{
		Title:     "Quote by " + quote.Author,
		Updated:   now,
		Published: now,
		ID:        quoteURL,
		Content:   quote.Text,
		Author:    AtomAuthor{Name: quote.Author},
		Link:      AtomLink{Href: quoteURL},
	}
}

func quoteToEmbeddedJS(quote ResponseQuote) string {
	return fmt.Sprintf(`
(function() {
    var quote = %s;
    var container = document.currentScript.parentElement;
    container.innerHTML = %s;
})();`,
		mustJSONString(quote),
		mustJSONString(quoteToEmbeddedHTML(quote)))
}

func mustJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func quoteToYAML(quote ResponseQuote) string {
	return fmt.Sprintf(`quote:
  id: %d
  text: "%s"
  author: "%s"
  tags:
    - %s
`, quote.ID, quote.Text, quote.Author, strings.Join(quote.Tags, "\n    - "))
}

func quoteToMarkdown(quote ResponseQuote) string {
	tags := strings.Join(quote.Tags, " #")
	if len(tags) > 0 {
		tags = "#" + tags
	}
	return fmt.Sprintf("> %s\n\n— %s\n\nQuote ID: %d\n\n%s", quote.Text, quote.Author, quote.ID, tags)
}

func quoteToRSS(quote ResponseQuote, quoteUrl string) string {
	pubDate := time.Now().Format(time.RFC1123Z)
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
  <title>Random Quote</title>
  <link>/quotes/%s</link>
  <description>Random Quote of the Moment</description>
  <item>
    <title>Quote by %s</title>
    <description>%s</description>
    <author>%s</author>
    <pubDate>%s</pubDate>
    <guid>%v</guid>
  </item>
</channel>
</rss>`, quoteUrl, quote.Author, quote.Text, quote.Author, pubDate, quoteUrl)
}

func quoteToHTML(quote ResponseQuote, highlightedTag string) string {
	var tagHTML strings.Builder
	for _, tag := range quote.Tags {
		cleanTag := strings.TrimSpace(tag)
		encodedTag := url.QueryEscape(cleanTag)
		tagClass := "tag"
		if cleanTag == highlightedTag {
			tagClass += " highlighted-tag"
		}
		tagHTML.WriteString(fmt.Sprintf(`<a href="/tags/%s" class="%s">%s</a>`, encodedTag, tagClass, cleanTag))
	}

	authorLink := fmt.Sprintf(`<a href="/authors/%s" class="author">— %s</a>`,
		strings.TrimSpace(quote.AuthorID),
		strings.TrimSpace(quote.Author))

	audioURL := fmt.Sprintf("/quote/%d?format=wav", quote.ID)
	quoteLink := fmt.Sprintf("/quote/%d", quote.ID)

	return fmt.Sprintf(`
    <div class="quote-container">
        <p class="quote-text">%s</p>
        %s
        <div class="tags">%s</div>
        <div class="footer">
            <div class="audio-container">
                <audio controls id="audio-%d" src="%s" preload="none">
                    Your browser does not support the audio element.
                </audio>
            </div>
            <div class="button-container">
                <a href="%s" class="quote-id">Quote ID: %d</a>
                <button class="icon-button share-button" onclick="shareQuote(event, '%s')" title="Share Quote">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="18" cy="5" r="3"></circle>
                        <circle cx="6" cy="12" r="3"></circle>
                        <circle cx="18" cy="19" r="3"></circle>
                        <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line>
                        <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>
                    </svg>
                </button>
                <button class="icon-button next-button" onclick="location.href='/random-quote'" title="Next Quote">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M23 4v6h-6"/>
                        <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                    </svg>
                </button>
            </div>
        </div>
    </div>
    <style>
        html {
            -webkit-text-size-adjust: 100%%;
            -ms-text-size-adjust: 100%%;
        }
        body {
            text-size-adjust: none;
            -webkit-text-size-adjust: none;
            -moz-text-size-adjust: none;
            -ms-text-size-adjust: none;
        }
        .quote-container {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica', 'Arial', sans-serif;
            max-width: 500px;
            margin: 20px auto;
            padding: 20px;
            background-color: #ffffff;
            border: 1px solid #e1e8ed;
            border-radius: 4px;
        }
        .quote-text {
            font-size: 18px !important;
            line-height: 1.5;
            color: #333333;
            margin-bottom: 15px;
            font-style: italic;
            max-width: 100%%;
            overflow-wrap: break-word;
            word-wrap: break-word;
        }
        .author {
            font-size: 16px !important;
            color: #657786;
            margin-bottom: 15px;
            text-decoration: none;
        }
        .author:hover {
            color: #1da1f2;
        }
        .tags {
            margin-bottom: 15px;
        }
        .tag {
            display: inline-block;
            background-color: #f1f3f5;
            color: #657786;
            padding: 4px 8px;
            border-radius: 16px;
            font-size: 14px !important;
            margin-right: 5px;
            margin-bottom: 5px;
            text-decoration: none;
        }
        .tag:hover {
            background-color: #e1e8ed;
            color: #1da1f2;
        }
        .highlighted-tag {
            background-color: #1da1f2 !important;
            color: white !important;
        }
        .highlighted-tag:hover {
            background-color: #1a91da !important;
        }
        .footer {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .quote-id {
            color: #aab8c2;
            font-size: 12px !important;
            text-decoration: none;
        }
        .quote-id:hover {
            color: #1da1f2;
        }
        .button-container {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .icon-button {
            background-color: transparent;
            color: #657786;
            border: none;
            padding: 4px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: color 0.3s ease;
        }
        .icon-button:hover {
            color: #1da1f2;
        }
        .share-button {
            position: relative;
        }
        .share-button::after {
            content: "Copied!";
            position: absolute;
            bottom: 100%%;
            left: 50%%;
            transform: translateX(-50%%);
            background-color: #657786;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            opacity: 0;
            transition: opacity 0.3s ease;
            pointer-events: none;
        }
        .share-button.copied::after {
            opacity: 1;
        }
        .audio-container {
            margin-top: 15px;
            margin-bottom: 15px;
        }
        .audio-container audio {
            height: 30px;
            opacity: 0.5;
            transition: opacity 0.3s ease;
            overflow: hidden;
            border-radius: 16px;
        }
        .audio-container audio:hover {
            opacity: 0.8;
        }
        .audio-container audio::-webkit-media-controls-enclosure {
            background-color: #f1f3f5;
            border-radius: 16px;
        }
        .audio-container audio::-webkit-media-controls-panel {
            background-color: transparent;
        }
        .audio-container audio::-webkit-media-controls-play-button {
            background-color: #657786;
            border-radius: 50%%;
        }
        .audio-container audio::-webkit-media-controls-current-time-display,
        .audio-container audio::-webkit-media-controls-time-remaining-display {
            color: #657786;
        }
        .audio-container audio::-webkit-media-controls-timeline {
            background-color: #e1e8ed;
            border-radius: 10px;
            margin-left: 10px;
            margin-right: 10px;
        }
        .audio-container audio::-webkit-media-controls-volume-slider,
        .audio-container audio::-webkit-media-controls-mute-button {
            display: none;
        }
        .social-share {
            display: flex;
            justify-content: space-between;
            margin-top: 15px;
        }
        .social-button {
            background-color: #f1f3f5;
            color: #657786;
            border: none;
            padding: 8px 12px;
            border-radius: 16px;
            font-size: 14px !important;
            cursor: pointer;
            transition: background-color 0.3s ease, color 0.3s ease;
        }
        .social-button:hover {
            background-color: #e1e8ed;
            color: #1da1f2;
        }
        @media screen and (max-width: 600px) {
            .quote-text {
                font-size: 16px !important;
            }
            .author {
                font-size: 14px !important;
            }
            .tag {
                font-size: 12px !important;
            }
            .audio-container audio {
                height: 25px;
            }
        }
    </style>
    <script>
   function shareQuote(event, quoteLink) {
    event.preventDefault();
    const shareUrl = window.location.origin + quoteLink;

    if (navigator.share) {
        navigator.share({
            title: 'Quote',
            text: 'Check out this quote!',
            url: shareUrl
        }).then(() => {
            console.log('Shared successfully');
        }).catch((error) => {
            console.log('Error sharing:', error);
            fallbackCopyToClipboard(shareUrl);
        });
    } else {
        fallbackCopyToClipboard(shareUrl);
    }
}

function fallbackCopyToClipboard(text) {
    const textArea = document.createElement("textarea");
    textArea.value = text;
    textArea.style.position = "fixed";  // Avoid scrolling to bottom
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
        const successful = document.execCommand('copy');
        const msg = successful ? 'successful' : 'unsuccessful';
        console.log('Fallback: Copying text command was ' + msg);
        showCopiedFeedback();
    } catch (err) {
        console.error('Fallback: Oops, unable to copy', err);
    }

    document.body.removeChild(textArea);
}

function showCopiedFeedback() {
    const shareButton = document.querySelector('.share-button');
    shareButton.classList.add('copied');
    setTimeout(() => {
        shareButton.classList.remove('copied');
    }, 2000);
}
    </script>
    `, quote.Text, authorLink, tagHTML.String(), quote.ID, audioURL, quoteLink, quote.ID, quoteLink)
}

func serveJSONQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)
}

func serveXMLQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "application/xml")
	xmlQuote := XMLQuote{
		Text:   q.Text,
		Author: q.Author,
		Tags:   q.Tags,
		ID:     q.ID,
	}
	xml.NewEncoder(w).Encode(xmlQuote)
}

func serveHTMLQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, quoteToHTML(q, ""))
}

func serveTextQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Quote: %s\nAuthor: %s\nTags: %v\nID: %d\n",
		q.Text, q.Author, strings.Join(q.Tags, ", "), q.ID)
}

func serveMarkdownQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "text/markdown")
	fmt.Fprint(w, quoteToMarkdown(q))
}

func serveYAMLQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "application/yaml")
	fmt.Fprint(w, quoteToYAML(q))
}

func serveCSVQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "text/csv")
	fmt.Fprint(w, quoteToCSV(q))
}

func serveRSSQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "application/rss+xml")
	fmt.Fprint(w, quoteToRSS(q, requestData.QuoteURL))
}

func serveAtomQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "application/atom+xml")
	feed := &AtomFeed{}
	feed.Create(q, requestData.BaseURL, requestData.QuoteURL)
	xml.NewEncoder(w).Encode(feed)
}

func serveOEmbedJSONQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "application/json+oembed")
	response := &OEmbedResponse{}
	response.Create(q, requestData.BaseURL)
	json.NewEncoder(w).Encode(response)
}

func serveOEmbedXMLQuote(w http.ResponseWriter, q ResponseQuote, api *API, requestData *ResponseInfo) {
	w.Header().Set("Content-Type", "text/xml+oembed")
	response := &OEmbedResponse{}
	response.Create(q, requestData.BaseURL)
	xml.NewEncoder(w).Encode(response)
}

func serveEmbedQuote(w http.ResponseWriter, q ResponseQuote, api *API, responseInfo *ResponseInfo) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, quoteToEmbeddedHTML(q))
}

func serveEmbedJSQuote(w http.ResponseWriter, q ResponseQuote, api *API, responseInfo *ResponseInfo) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, quoteToEmbeddedJS(q))
}

func serveSVGQuote(w http.ResponseWriter, q ResponseQuote, api *API, responseInfo *ResponseInfo) {
	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, quoteToSVG(q))
}
func serveSVGQuoteDownload(w http.ResponseWriter, q ResponseQuote, api *API, responseInfo *ResponseInfo) {
	svgContent := quoteToSVG(q)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"quote_%d.svg\"", q.ID))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(svgContent)))

	_, err := w.Write([]byte(svgContent))
	if err != nil {
		http.Error(w, "Error writing SVG content", http.StatusInternalServerError)
		return
	}
}
