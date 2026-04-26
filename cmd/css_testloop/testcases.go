package main

// CSSTestCase defines a single visual test for one CSS property/feature.
// Each test case produces a small cell (CellW x CellH) rendered both
// in HTML/CSS (reference) and in ebitenui-xml (actual).
type CSSTestCase struct {
	ID       string // unique, e.g. "bg-solid"
	Category string // "background", "border", "layout", ...
	Property string // CSS property being tested
	Label    string // short human label

	// ebitenui-xml side
	XML    string // XML layout snippet (root must be <panel>)
	Styles string // JSON styles (flat format)

	// HTML/CSS reference side
	HTML string // HTML snippet
	CSS  string // CSS rules
}

const (
	CellW = 200
	CellH = 150
)

func AllTestCases() []CSSTestCase {
	return []CSSTestCase{
		// ── Background ──
		{
			ID: "bg-solid", Category: "background", Property: "background",
			Label:  "Solid background",
			XML:    `<panel id="t"><text>BG</text></panel>`,
			Styles: `{"#t":{"background":"#e74c3c","width":180,"height":130,"padding":{"all":10},"direction":"column","align":"center","justify":"center"},"text":{"color":"#fff"}}`,
			HTML:   `<div id="t"><span>BG</span></div>`,
			CSS:    `#t{background:#e74c3c;width:180px;height:130px;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:10px;box-sizing:border-box} #t span{color:#fff;font-family:monospace}`,
		},
		{
			ID: "bg-gradient-h", Category: "background", Property: "background(gradient)",
			Label:  "Horizontal gradient",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"background":"linear-gradient(90deg, #e74c3c, #3498db)","width":180,"height":130}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{background:linear-gradient(90deg,#e74c3c,#3498db);width:180px;height:130px}`,
		},
		{
			ID: "bg-gradient-v", Category: "background", Property: "background(gradient-v)",
			Label:  "Vertical gradient",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"background":"linear-gradient(180deg, #2ecc71, #8e44ad)","width":180,"height":130}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{background:linear-gradient(180deg,#2ecc71,#8e44ad);width:180px;height:130px}`,
		},
		{
			ID: "bg-gradient-diag", Category: "background", Property: "background(gradient-45)",
			Label:  "45deg gradient",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"background":"linear-gradient(45deg, #f39c12, #2980b9)","width":180,"height":130}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{background:linear-gradient(45deg,#f39c12,#2980b9);width:180px;height:130px}`,
		},

		// ── Border ──
		{
			ID: "border-basic", Category: "border", Property: "border",
			Label:  "Basic border",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#1a1a2e","borderWidth":3,"border":"#e74c3c","borderRadius":0}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:180px;height:130px;background:#1a1a2e;border:3px solid #e74c3c;box-sizing:border-box}`,
		},
		{
			ID: "border-radius", Category: "border", Property: "border-radius",
			Label:  "Rounded corners",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#3498db","borderRadius":20}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:180px;height:130px;background:#3498db;border-radius:20px}`,
		},
		{
			ID: "border-radius-per-corner", Category: "border", Property: "border-radius(per-corner)",
			Label:  "Per-corner radius",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#9b59b6","borderTopLeftRadius":30,"borderTopRightRadius":0,"borderBottomRightRadius":30,"borderBottomLeftRadius":0}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:180px;height:130px;background:#9b59b6;border-radius:30px 0 30px 0}`,
		},

		// ── Box Shadow ──
		{
			ID: "box-shadow", Category: "shadow", Property: "box-shadow",
			Label:  "Box shadow",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":160,"height":110,"background":"#ecf0f1","borderRadius":8,"boxShadow":"4 4 12 0 rgba(0,0,0,0.5)","margin":{"all":15}}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:160px;height:110px;background:#ecf0f1;border-radius:8px;box-shadow:4px 4px 12px 0px rgba(0,0,0,0.5);margin:15px}`,
		},

		// ── Layout: flex-direction ──
		{
			ID: "flex-row", Category: "layout", Property: "flex-direction:row",
			Label:  "Flex row",
			XML:    `<panel id="t"><panel class="a"/><panel class="b"/><panel class="c"/></panel>`,
			Styles: `{"#t":{"direction":"row","gap":8,"width":180,"height":130,"padding":{"all":10}},"panel":{"width":40,"height":40},".a":{"background":"#e74c3c"},".b":{"background":"#3498db"},".c":{"background":"#2ecc71"}}`,
			HTML:   `<div id="t"><div class="a"></div><div class="b"></div><div class="c"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:row;gap:8px;width:180px;height:130px;padding:10px;box-sizing:border-box} .a{width:40px;height:40px;background:#e74c3c} .b{width:40px;height:40px;background:#3498db} .c{width:40px;height:40px;background:#2ecc71}`,
		},
		{
			ID: "flex-col", Category: "layout", Property: "flex-direction:column",
			Label:  "Flex column",
			XML:    `<panel id="t"><panel class="a"/><panel class="b"/><panel class="c"/></panel>`,
			Styles: `{"#t":{"direction":"column","gap":6,"width":180,"height":130,"padding":{"all":10}},"panel":{"height":25},".a":{"background":"#e74c3c"},".b":{"background":"#3498db"},".c":{"background":"#2ecc71"}}`,
			HTML:   `<div id="t"><div class="a"></div><div class="b"></div><div class="c"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:column;gap:6px;width:180px;height:130px;padding:10px;box-sizing:border-box} .a{height:25px;background:#e74c3c} .b{height:25px;background:#3498db} .c{height:25px;background:#2ecc71}`,
		},

		// ── justify-content ──
		{
			ID: "justify-center", Category: "layout", Property: "justify-content:center",
			Label:  "Justify center",
			XML:    `<panel id="t"><panel class="a"/><panel class="b"/></panel>`,
			Styles: `{"#t":{"direction":"row","justify":"center","width":180,"height":130,"background":"#2c3e50","padding":{"all":10}},".a":{"width":30,"height":30,"background":"#e74c3c"},".b":{"width":30,"height":30,"background":"#3498db"}}`,
			HTML:   `<div id="t"><div class="a"></div><div class="b"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:row;justify-content:center;width:180px;height:130px;background:#2c3e50;padding:10px;box-sizing:border-box} .a{width:30px;height:30px;background:#e74c3c} .b{width:30px;height:30px;background:#3498db}`,
		},
		{
			ID: "justify-between", Category: "layout", Property: "justify-content:space-between",
			Label:  "Space between",
			XML:    `<panel id="t"><panel class="a"/><panel class="b"/><panel class="c"/></panel>`,
			Styles: `{"#t":{"direction":"row","justify":"space-between","width":180,"height":130,"background":"#2c3e50","padding":{"all":10}},".a":{"width":30,"height":30,"background":"#e74c3c"},".b":{"width":30,"height":30,"background":"#3498db"},".c":{"width":30,"height":30,"background":"#2ecc71"}}`,
			HTML:   `<div id="t"><div class="a"></div><div class="b"></div><div class="c"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:row;justify-content:space-between;width:180px;height:130px;background:#2c3e50;padding:10px;box-sizing:border-box} .a{width:30px;height:30px;background:#e74c3c} .b{width:30px;height:30px;background:#3498db} .c{width:30px;height:30px;background:#2ecc71}`,
		},
		{
			ID: "justify-evenly", Category: "layout", Property: "justify-content:space-evenly",
			Label:  "Space evenly",
			XML:    `<panel id="t"><panel class="a"/><panel class="b"/><panel class="c"/></panel>`,
			Styles: `{"#t":{"direction":"row","justify":"space-evenly","width":180,"height":130,"background":"#2c3e50","padding":{"all":10}},".a":{"width":30,"height":30,"background":"#e74c3c"},".b":{"width":30,"height":30,"background":"#3498db"},".c":{"width":30,"height":30,"background":"#2ecc71"}}`,
			HTML:   `<div id="t"><div class="a"></div><div class="b"></div><div class="c"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:row;justify-content:space-evenly;width:180px;height:130px;background:#2c3e50;padding:10px;box-sizing:border-box} .a{width:30px;height:30px;background:#e74c3c} .b{width:30px;height:30px;background:#3498db} .c{width:30px;height:30px;background:#2ecc71}`,
		},

		// ── align-items ──
		{
			ID: "align-center", Category: "layout", Property: "align-items:center",
			Label:  "Align center",
			XML:    `<panel id="t"><panel class="a"/><panel class="b"/></panel>`,
			Styles: `{"#t":{"direction":"row","align":"center","width":180,"height":130,"background":"#34495e","padding":{"all":10}},".a":{"width":30,"height":20,"background":"#e74c3c"},".b":{"width":30,"height":50,"background":"#3498db"}}`,
			HTML:   `<div id="t"><div class="a"></div><div class="b"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:row;align-items:center;width:180px;height:130px;background:#34495e;padding:10px;box-sizing:border-box} .a{width:30px;height:20px;background:#e74c3c} .b{width:30px;height:50px;background:#3498db}`,
		},

		// ── flexGrow ──
		{
			ID: "flex-grow", Category: "layout", Property: "flex-grow",
			Label:  "Flex grow",
			XML:    `<panel id="t"><panel class="a"/><panel class="b"/></panel>`,
			Styles: `{"#t":{"direction":"row","width":180,"height":130,"background":"#2c3e50","padding":{"all":10}},".a":{"flexGrow":1,"height":40,"background":"#e74c3c"},".b":{"flexGrow":2,"height":40,"background":"#3498db"}}`,
			HTML:   `<div id="t"><div class="a"></div><div class="b"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:row;width:180px;height:130px;background:#2c3e50;padding:10px;box-sizing:border-box} .a{flex-grow:1;height:40px;background:#e74c3c} .b{flex-grow:2;height:40px;background:#3498db}`,
		},

		// ── Padding ──
		{
			ID: "padding", Category: "spacing", Property: "padding",
			Label:  "Padding",
			XML:    `<panel id="t"><panel class="inner"/></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#2c3e50","padding":{"top":20,"right":10,"bottom":20,"left":10},"direction":"column"},".inner":{"background":"#e74c3c","flexGrow":1}}`,
			HTML:   `<div id="t"><div class="inner"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:column;width:180px;height:130px;background:#2c3e50;padding:20px 10px;box-sizing:border-box} .inner{flex-grow:1;background:#e74c3c}`,
		},

		// ── Opacity ──
		{
			ID: "opacity", Category: "visual", Property: "opacity",
			Label:  "Opacity 0.5",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#e74c3c","opacity":0.5}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:180px;height:130px;background:#e74c3c;opacity:0.5}`,
		},

		// ── Nested layout ──
		{
			ID: "nested-layout", Category: "layout", Property: "nested flex",
			Label:  "Nested flex",
			XML:    `<panel id="t"><panel class="top"><panel class="a"/><panel class="b"/></panel><panel class="bot"/></panel>`,
			Styles: `{"#t":{"direction":"column","width":180,"height":130,"gap":6,"background":"#1a1a2e","padding":{"all":8}},".top":{"direction":"row","gap":6,"flexGrow":1},".a":{"flexGrow":1,"background":"#e74c3c"},".b":{"flexGrow":1,"background":"#3498db"},".bot":{"height":30,"background":"#2ecc71"}}`,
			HTML:   `<div id="t"><div class="top"><div class="a"></div><div class="b"></div></div><div class="bot"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:column;width:180px;height:130px;gap:6px;background:#1a1a2e;padding:8px;box-sizing:border-box} .top{display:flex;flex-direction:row;gap:6px;flex-grow:1} .a{flex-grow:1;background:#e74c3c} .b{flex-grow:1;background:#3498db} .bot{height:30px;background:#2ecc71}`,
		},

		// ── Overflow hidden ──
		{
			ID: "overflow-hidden", Category: "visual", Property: "overflow:hidden",
			Label:  "Overflow hidden",
			XML:    `<panel id="t"><panel class="big"/></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#2c3e50","overflow":"hidden","direction":"column"},".big":{"width":300,"height":300,"background":"#e74c3c"}}`,
			HTML:   `<div id="t"><div class="big"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:column;width:180px;height:130px;background:#2c3e50;overflow:hidden} .big{width:300px;height:300px;background:#e74c3c;flex-shrink:0}`,
		},
		{
			ID: "overflow-scroll", Category: "visual", Property: "overflow:scroll",
			Label:  "Overflow scroll",
			XML:    `<panel id="t"><panel class="a"/><panel class="b"/><panel class="c"/></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#263238","overflow":"scroll","direction":"column"},".a":{"height":60,"background":"#e74c3c","flexShrink":0},".b":{"height":60,"background":"#3498db","flexShrink":0},".c":{"height":60,"background":"#2ecc71","flexShrink":0}}`,
			HTML:   `<div id="t"><div class="a"></div><div class="b"></div><div class="c"></div></div>`,
			CSS:    `#t{display:flex;flex-direction:column;width:180px;height:130px;background:#263238;overflow:scroll} .a{height:60px;background:#e74c3c;flex-shrink:0}.b{height:60px;background:#3498db;flex-shrink:0}.c{height:60px;background:#2ecc71;flex-shrink:0}`,
		},

		// ── Border + background combined ──
		{
			ID: "position-absolute", Category: "layout", Property: "position:absolute",
			Label:  "Absolute inset",
			XML:    `<panel id="t"><panel class="flow"/><panel class="abs"/></panel>`,
			Styles: `{"#t":{"direction":"row","width":180,"height":130,"background":"#17202a"},".flow":{"width":45,"height":45,"background":"#2ecc71"},".abs":{"position":"absolute","left":90,"top":35,"width":55,"height":45,"background":"#e74c3c"}}`,
			HTML:   `<div id="t"><div class="flow"></div><div class="abs"></div></div>`,
			CSS:    `#t{position:relative;display:flex;flex-direction:row;width:180px;height:130px;background:#17202a}.flow{width:45px;height:45px;background:#2ecc71}.abs{position:absolute;left:90px;top:35px;width:55px;height:45px;background:#e74c3c}`,
		},
		{
			ID: "z-index-overlap", Category: "visual", Property: "z-index",
			Label:  "Z overlap",
			XML:    `<panel id="t"><panel class="low"/><panel class="high"/></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#ecf0f1"},".low":{"position":"absolute","left":45,"top":35,"width":70,"height":55,"background":"#3498db","zIndex":1},".high":{"position":"absolute","left":75,"top":55,"width":70,"height":55,"background":"#e74c3c","zIndex":5}}`,
			HTML:   `<div id="t"><div class="low"></div><div class="high"></div></div>`,
			CSS:    `#t{position:relative;width:180px;height:130px;background:#ecf0f1}.low{position:absolute;left:45px;top:35px;width:70px;height:55px;background:#3498db;z-index:1}.high{position:absolute;left:75px;top:55px;width:70px;height:55px;background:#e74c3c;z-index:5}`,
		},
		{
			ID: "multi-shadow", Category: "shadow", Property: "box-shadow(list)",
			Label:  "Multi shadow",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":140,"height":90,"margin":{"all":25},"background":"#f8f9fa","borderRadius":12,"boxShadow":"0 8px 18px 0 rgba(0,0,0,0.35), 0 0 0 4px rgba(52,152,219,0.45)"}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:140px;height:90px;margin:25px;background:#f8f9fa;border-radius:12px;box-shadow:0 8px 18px 0 rgba(0,0,0,0.35),0 0 0 4px rgba(52,152,219,0.45)}`,
		},
		{
			ID: "box-shadow-corner-radii", Category: "shadow", Property: "box-shadow(per-corner-radius)",
			Label:  "Per-corner shadow",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":150,"height":95,"margin":{"all":25},"background":"#fdfefe","borderTopLeftRadius":32,"borderTopRightRadius":4,"borderBottomRightRadius":32,"borderBottomLeftRadius":4,"boxShadow":"0 8px 18px 0 rgba(0,0,0,0.38)"}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:150px;height:95px;margin:25px;background:#fdfefe;border-top-left-radius:32px;border-top-right-radius:4px;border-bottom-right-radius:32px;border-bottom-left-radius:4px;box-shadow:0 8px 18px 0 rgba(0,0,0,0.38)}`,
		},
		{
			ID: "text-shadow-blur", Category: "shadow", Property: "text-shadow(blur)",
			Label:  "Text shadow blur",
			XML:    `<panel id="t"><text class="label">Glow</text></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#1b2631","direction":"column","align":"center","justify":"center"},".label":{"color":"#fff","fontSize":24,"textShadow":"0 0 8px rgba(46,204,113,0.9), 2px 2px 3px rgba(0,0,0,0.6)"}}`,
			HTML:   `<div id="t"><span>Glow</span></div>`,
			CSS:    `#t{width:180px;height:130px;background:#1b2631;display:flex;align-items:center;justify-content:center}span{color:#fff;font:24px monospace;text-shadow:0 0 8px rgba(46,204,113,0.9),2px 2px 3px rgba(0,0,0,0.6)}`,
		},
		{
			ID: "clip-path-inset", Category: "visual", Property: "clip-path:inset",
			Label:  "Clip inset",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"linear-gradient(90deg, #e74c3c, #3498db)","clipPath":"inset(18px 28px)"}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:180px;height:130px;background:linear-gradient(90deg,#e74c3c,#3498db);clip-path:inset(18px 28px)}`,
		},
		{
			ID: "clip-path-circle", Category: "visual", Property: "clip-path:circle",
			Label:  "Clip circle",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"linear-gradient(135deg, #f1c40f, #8e44ad)","clipPath":"circle(42% at 50% 50%)"}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:180px;height:130px;background:linear-gradient(135deg,#f1c40f,#8e44ad);clip-path:circle(42% at 50% 50%)}`,
		},
		{
			ID: "clip-path-polygon", Category: "visual", Property: "clip-path:polygon",
			Label:  "Clip polygon",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"linear-gradient(135deg, #16a085, #f39c12)","clipPath":"polygon(50% 0, 100% 100%, 0 100%)"}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:180px;height:130px;background:linear-gradient(135deg,#16a085,#f39c12);clip-path:polygon(50% 0,100% 100%,0 100%)}`,
		},
		{
			ID: "clip-path-path", Category: "visual", Property: "clip-path:path",
			Label:  "Clip path",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"linear-gradient(135deg, #2ecc71, #9b59b6)","clipPath":"path('M0 0 L180 0 L150 130 L30 130 Z')"}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:180px;height:130px;background:linear-gradient(135deg,#2ecc71,#9b59b6);clip-path:path('M0 0 L180 0 L150 130 L30 130 Z')}`,
		},
		{
			ID: "border-bg-combined", Category: "border", Property: "border+bg+radius",
			Label:  "Border+BG+Radius",
			XML:    `<panel id="t"></panel>`,
			Styles: `{"#t":{"width":180,"height":130,"background":"#1abc9c","borderWidth":4,"border":"#f39c12","borderRadius":16}}`,
			HTML:   `<div id="t"></div>`,
			CSS:    `#t{width:180px;height:130px;background:#1abc9c;border:4px solid #f39c12;border-radius:16px;box-sizing:border-box}`,
		},
	}
}
