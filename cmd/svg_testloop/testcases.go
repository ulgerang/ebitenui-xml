package main

// SVGTestCase defines a single SVG rendering test.
// The SVG string is used identically for both Ebiten (ParseSVGString) and HTML (<svg>).
type SVGTestCase struct {
	ID       string // unique test ID
	Category string // "shape", "path", "style", "transform", "icon"
	Label    string // short description
	SVG      string // SVG markup (used by BOTH Go and HTML)
}

const (
	CellW = 200
	CellH = 200
)

func AllTestCases() []SVGTestCase {
	return []SVGTestCase{
		// ── Basic Shapes ──

		{
			ID: "rect-basic", Category: "shape", Label: "Basic rect",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="10" y="10" width="80" height="80" fill="#e74c3c"/>
			</svg>`,
		},
		{
			ID: "rect-rounded", Category: "shape", Label: "Rounded rect",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="10" y="10" width="80" height="80" rx="15" ry="15" fill="#3498db"/>
			</svg>`,
		},
		{
			ID: "rect-stroke", Category: "shape", Label: "Rect with stroke",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="10" y="10" width="80" height="80" fill="#2ecc71" stroke="#e74c3c" stroke-width="4"/>
			</svg>`,
		},
		{
			ID: "circle-basic", Category: "shape", Label: "Basic circle",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<circle cx="50" cy="50" r="40" fill="#9b59b6"/>
			</svg>`,
		},
		{
			ID: "circle-stroke", Category: "shape", Label: "Circle with stroke",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<circle cx="50" cy="50" r="35" fill="none" stroke="#f39c12" stroke-width="5"/>
			</svg>`,
		},
		{
			ID: "ellipse-basic", Category: "shape", Label: "Ellipse",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<ellipse cx="50" cy="50" rx="45" ry="25" fill="#1abc9c"/>
			</svg>`,
		},
		{
			ID: "line-basic", Category: "shape", Label: "Lines",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<line x1="10" y1="10" x2="90" y2="90" stroke="#e74c3c" stroke-width="3"/>
				<line x1="90" y1="10" x2="10" y2="90" stroke="#3498db" stroke-width="3"/>
			</svg>`,
		},

		// ── Polygon & Polyline ──

		{
			ID: "polygon-triangle", Category: "shape", Label: "Polygon triangle",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<polygon points="50,10 90,90 10,90" fill="#e74c3c" stroke="#f39c12" stroke-width="2"/>
			</svg>`,
		},
		{
			ID: "polygon-star", Category: "shape", Label: "Polygon star",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<polygon points="50,5 61,35 95,35 68,57 79,91 50,70 21,91 32,57 5,35 39,35" fill="#f39c12" stroke="#e67e22" stroke-width="1"/>
			</svg>`,
		},
		{
			ID: "polyline-zigzag", Category: "shape", Label: "Polyline zigzag",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<polyline points="10,80 25,20 40,60 55,15 70,50 85,10" fill="none" stroke="#2ecc71" stroke-width="3"/>
			</svg>`,
		},

		// ── SVG Path Commands ──

		{
			ID: "path-lines", Category: "path", Label: "Path M/L/Z",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M10,10 L90,10 L90,90 L10,90 Z" fill="#3498db" stroke="#2c3e50" stroke-width="2"/>
			</svg>`,
		},
		{
			ID: "path-hv", Category: "path", Label: "Path H/V",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M10,10 H90 V50 H50 V90 H10 Z" fill="#e74c3c" stroke="none"/>
			</svg>`,
		},
		{
			ID: "path-cubic", Category: "path", Label: "Path Cubic Bezier",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M10,80 C10,10 90,10 90,80" fill="none" stroke="#9b59b6" stroke-width="3"/>
			</svg>`,
		},
		{
			ID: "path-quad", Category: "path", Label: "Path Quad Bezier",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M10,90 Q50,10 90,90" fill="none" stroke="#e67e22" stroke-width="3"/>
			</svg>`,
		},
		{
			ID: "path-arc", Category: "path", Label: "Path Arc",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M20,50 A30,30 0 1,1 80,50 A30,30 0 1,1 20,50 Z" fill="#1abc9c" stroke="none"/>
			</svg>`,
		},
		{
			ID: "path-smooth-cubic", Category: "path", Label: "Path Smooth S",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M10,50 C10,10 40,10 50,50 S90,90 90,50" fill="none" stroke="#3498db" stroke-width="3"/>
			</svg>`,
		},
		{
			ID: "path-relative", Category: "path", Label: "Relative cmds",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M10,50 l20,-30 l20,30 l20,-30 l20,30" fill="none" stroke="#e74c3c" stroke-width="3"/>
			</svg>`,
		},

		// ── Styling ──

		{
			ID: "style-opacity", Category: "style", Label: "Opacity",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="10" y="10" width="60" height="60" fill="#e74c3c"/>
				<rect x="30" y="30" width="60" height="60" fill="#3498db" opacity="0.5"/>
			</svg>`,
		},
		{
			ID: "style-fill-none", Category: "style", Label: "Fill none (stroke only)",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<circle cx="50" cy="50" r="35" fill="none" stroke="#e74c3c" stroke-width="4"/>
				<rect x="25" y="25" width="50" height="50" fill="none" stroke="#3498db" stroke-width="3"/>
			</svg>`,
		},
		{
			ID: "style-multi-shapes", Category: "style", Label: "Multiple shapes",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="5" y="5" width="40" height="40" fill="#e74c3c"/>
				<rect x="55" y="5" width="40" height="40" fill="#3498db"/>
				<circle cx="25" cy="75" r="20" fill="#2ecc71"/>
				<circle cx="75" cy="75" r="20" fill="#f39c12"/>
			</svg>`,
		},

		// ── Groups & Transform ──

		{
			ID: "group-translate", Category: "transform", Label: "Group translate",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<g transform="translate(20,20)">
					<rect x="0" y="0" width="40" height="40" fill="#e74c3c"/>
					<circle cx="20" cy="20" r="10" fill="#f1c40f"/>
				</g>
			</svg>`,
		},
		{
			ID: "group-scale", Category: "transform", Label: "Group scale",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<g transform="scale(0.5)">
					<rect x="10" y="10" width="80" height="80" fill="#3498db"/>
				</g>
				<g transform="translate(50,50) scale(0.5)">
					<rect x="10" y="10" width="80" height="80" fill="#e74c3c"/>
				</g>
			</svg>`,
		},
		{
			ID: "group-nested", Category: "transform", Label: "Nested groups",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<g fill="#3498db">
					<rect x="5" y="5" width="30" height="30"/>
					<g fill="#e74c3c">
						<rect x="40" y="5" width="30" height="30"/>
						<circle cx="55" cy="55" r="15"/>
					</g>
					<rect x="75" y="5" width="20" height="20"/>
				</g>
			</svg>`,
		},
		{
			ID: "group-inherit", Category: "transform", Label: "Inherited styles",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<g fill="none" stroke="#e74c3c" stroke-width="3">
					<circle cx="30" cy="30" r="20"/>
					<circle cx="70" cy="30" r="20"/>
					<circle cx="50" cy="70" r="20"/>
				</g>
			</svg>`,
		},

		// ── CommonIcons (path-based icons) ──

		{
			ID: "icon-heart", Category: "icon", Label: "Heart icon",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24">
				<path d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z" fill="none" stroke="#E91E63" stroke-width="2"/>
			</svg>`,
		},
		{
			ID: "icon-star", Category: "icon", Label: "Star icon",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24">
				<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" fill="none" stroke="#FFC107" stroke-width="2"/>
			</svg>`,
		},
		{
			ID: "icon-check", Category: "icon", Label: "Check icon",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24">
				<path d="M20 6L9 17l-5-5" fill="none" stroke="#4CAF50" stroke-width="2.5"/>
			</svg>`,
		},
		{
			ID: "icon-search", Category: "icon", Label: "Search icon",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24">
				<path d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" fill="none" stroke="#2196F3" stroke-width="2"/>
			</svg>`,
		},
		{
			ID: "icon-home", Category: "icon", Label: "Home icon",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24">
				<path d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" fill="none" stroke="#8BC34A" stroke-width="2"/>
			</svg>`,
		},

		// ── ViewBox / Scaling ──

		{
			ID: "viewbox-scale-up", Category: "viewbox", Label: "ViewBox scale up",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 50 50" width="50" height="50">
				<circle cx="25" cy="25" r="20" fill="#e74c3c"/>
				<rect x="15" y="15" width="20" height="20" fill="#f1c40f" opacity="0.8"/>
			</svg>`,
		},
		{
			ID: "viewbox-offset", Category: "viewbox", Label: "ViewBox offset",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="10 10 80 80" width="100" height="100">
				<rect x="10" y="10" width="80" height="80" fill="#2c3e50"/>
				<circle cx="50" cy="50" r="30" fill="#e74c3c"/>
			</svg>`,
		},

		// ── Complex Composition ──

		{
			ID: "complex-face", Category: "complex", Label: "Simple face",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<circle cx="50" cy="50" r="45" fill="#f1c40f" stroke="#e67e22" stroke-width="2"/>
				<circle cx="35" cy="40" r="5" fill="#2c3e50"/>
				<circle cx="65" cy="40" r="5" fill="#2c3e50"/>
				<path d="M30,65 Q50,85 70,65" fill="none" stroke="#2c3e50" stroke-width="3"/>
			</svg>`,
		},
		{
			ID: "complex-badge", Category: "complex", Label: "Badge shape",
			SVG: `			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<circle cx="50" cy="50" r="40" fill="#3498db"/>
				<circle cx="50" cy="50" r="32" fill="none" stroke="#ecf0f1" stroke-width="2"/>
				<path d="M50 25 l5 10 l11 2 l-8 8 l2 11 l-10-5 l-10 5 l2-11 l-8-8 l11-2 z" fill="#f1c40f"/>
			</svg>`,
		},

		// ── Phase 1-2 New Tests ──

		{
			ID: "group-deep-nested", Category: "transform", Label: "Deep nested groups",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<g fill="#3498db">
					<rect x="5" y="5" width="20" height="20"/>
					<g fill="#e74c3c">
						<rect x="30" y="5" width="20" height="20"/>
						<g fill="#2ecc71">
							<rect x="55" y="5" width="20" height="20"/>
							<circle cx="65" cy="50" r="12"/>
						</g>
						<circle cx="40" cy="50" r="12"/>
					</g>
					<rect x="80" y="5" width="15" height="20"/>
				</g>
			</svg>`,
		},
		{
			ID: "transform-rotate", Category: "transform", Label: "Rotate 45°",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<g transform="rotate(45, 50, 50)">
					<rect x="25" y="25" width="50" height="50" fill="#3498db"/>
				</g>
			</svg>`,
		},
		{
			ID: "transform-rotate-origin", Category: "transform", Label: "Rotate with origin",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="30" y="30" width="40" height="40" fill="#bdc3c7"/>
				<g transform="rotate(30, 50, 50)">
					<rect x="30" y="30" width="40" height="40" fill="#e74c3c" opacity="0.8"/>
				</g>
			</svg>`,
		},
		{
			ID: "fill-rule-evenodd", Category: "style", Label: "fill-rule evenodd",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M50,5 L61,40 L98,40 L68,62 L79,97 L50,75 L21,97 L32,62 L2,40 L39,40 Z"
				      fill="#e74c3c" fill-rule="evenodd" stroke="#2c3e50" stroke-width="1"/>
			</svg>`,
		},
		{
			ID: "style-linecap", Category: "style", Label: "Stroke linecap",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<line x1="15" y1="20" x2="85" y2="20" stroke="#e74c3c" stroke-width="10" stroke-linecap="butt"/>
				<line x1="15" y1="50" x2="85" y2="50" stroke="#3498db" stroke-width="10" stroke-linecap="round"/>
				<line x1="15" y1="80" x2="85" y2="80" stroke="#2ecc71" stroke-width="10" stroke-linecap="square"/>
			</svg>`,
		},

		// ── Phase 3-4 New Tests ──

		{
			ID: "style-stroke-opacity", Category: "style", Label: "Stroke opacity",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="10" y="10" width="80" height="80" fill="#3498db" stroke="#e74c3c" stroke-width="8" stroke-opacity="0.4"/>
			</svg>`,
		},
		{
			ID: "style-fill-opacity", Category: "style", Label: "Fill opacity",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="10" y="10" width="60" height="60" fill="#e74c3c" fill-opacity="0.3"/>
				<circle cx="60" cy="60" r="30" fill="#3498db" fill-opacity="0.5" stroke="#2c3e50" stroke-width="2"/>
			</svg>`,
		},
		{
			ID: "style-linejoin", Category: "style", Label: "Stroke linejoin",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<polyline points="10,30 50,10 90,30" fill="none" stroke="#e74c3c" stroke-width="6" stroke-linejoin="miter"/>
				<polyline points="10,60 50,40 90,60" fill="none" stroke="#3498db" stroke-width="6" stroke-linejoin="round"/>
				<polyline points="10,90 50,70 90,90" fill="none" stroke="#2ecc71" stroke-width="6" stroke-linejoin="bevel"/>
			</svg>`,
		},
		{
			ID: "style-dasharray", Category: "style", Label: "Stroke dasharray",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<line x1="10" y1="20" x2="90" y2="20" stroke="#e74c3c" stroke-width="3" stroke-dasharray="10,5"/>
				<line x1="10" y1="50" x2="90" y2="50" stroke="#3498db" stroke-width="3" stroke-dasharray="5,5"/>
				<line x1="10" y1="80" x2="90" y2="80" stroke="#2ecc71" stroke-width="3" stroke-dasharray="15,5,5,5"/>
			</svg>`,
		},
		{
			ID: "transform-skew", Category: "transform", Label: "SkewX/SkewY",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="20" y="20" width="30" height="30" fill="#bdc3c7"/>
				<g transform="skewX(20)">
					<rect x="20" y="20" width="30" height="30" fill="#e74c3c" opacity="0.7"/>
				</g>
				<g transform="translate(50,0) skewY(15)">
					<rect x="0" y="20" width="30" height="30" fill="#3498db" opacity="0.7"/>
				</g>
			</svg>`,
		},
		{
			ID: "transform-matrix", Category: "transform", Label: "Matrix transform",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="20" y="20" width="40" height="40" fill="#bdc3c7"/>
				<g transform="matrix(0.866,0.5,-0.5,0.866,30,-10)">
					<rect x="20" y="20" width="40" height="40" fill="#9b59b6" opacity="0.8"/>
				</g>
			</svg>`,
		},

		// ── Phase 5-1 Text Tests ──

		{
			ID: "text-basic", Category: "text", Label: "Basic text",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<text x="10" y="30" font-size="20" fill="#e74c3c">Hello</text>
				<text x="10" y="60" font-size="14" fill="#3498db">SVG Text</text>
				<text x="10" y="85" font-size="10" fill="#2ecc71">Small</text>
			</svg>`,
		},
		{
			ID: "text-anchor", Category: "text", Label: "Text anchor",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<line x1="50" y1="0" x2="50" y2="100" stroke="#cccccc" stroke-width="1"/>
				<text x="50" y="25" font-size="14" fill="#e74c3c" text-anchor="start">start</text>
				<text x="50" y="50" font-size="14" fill="#3498db" text-anchor="middle">middle</text>
				<text x="50" y="75" font-size="14" fill="#2ecc71" text-anchor="end">end</text>
			</svg>`,
		},

		// ── Phase 5-2 Gradient Tests ──

		{
			ID: "gradient-linear", Category: "gradient", Label: "Linear gradient",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<defs>
					<linearGradient id="lr" x1="0" y1="0" x2="1" y2="0">
						<stop offset="0%" stop-color="#e74c3c"/>
						<stop offset="100%" stop-color="#3498db"/>
					</linearGradient>
					<linearGradient id="tb" x1="0" y1="0" x2="0" y2="1">
						<stop offset="0%" stop-color="#2ecc71"/>
						<stop offset="100%" stop-color="#f39c12"/>
					</linearGradient>
					<linearGradient id="diag" x1="0" y1="0" x2="1" y2="1">
						<stop offset="0%" stop-color="#9b59b6"/>
						<stop offset="50%" stop-color="#1abc9c"/>
						<stop offset="100%" stop-color="#e67e22"/>
					</linearGradient>
				</defs>
				<rect x="5" y="5" width="40" height="25" fill="url(#lr)"/>
				<rect x="55" y="5" width="40" height="25" fill="url(#tb)"/>
				<rect x="5" y="40" width="40" height="25" fill="url(#diag)"/>
				<circle cx="75" cy="52" r="12" fill="url(#lr)"/>
				<path d="M 10 75 L 40 75 L 25 95 Z" fill="url(#tb)"/>
			</svg>`,
		},
		{
			ID: "gradient-radial", Category: "gradient", Label: "Radial gradient",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<defs>
					<radialGradient id="rg1">
						<stop offset="0%" stop-color="#f39c12"/>
						<stop offset="100%" stop-color="#e74c3c"/>
					</radialGradient>
					<radialGradient id="rg2" cx="0.3" cy="0.3" r="0.5">
						<stop offset="0%" stop-color="#ffffff"/>
						<stop offset="100%" stop-color="#3498db"/>
					</radialGradient>
				</defs>
				<rect x="5" y="5" width="40" height="40" fill="url(#rg1)"/>
				<circle cx="75" cy="25" r="20" fill="url(#rg2)"/>
				<ellipse cx="25" cy="75" rx="20" ry="15" fill="url(#rg1)"/>
				<path d="M 55 60 L 95 60 L 95 95 L 55 95 Z" fill="url(#rg2)"/>
			</svg>`,
		},

		// ── Use element ──

		{
			ID: "use-basic", Category: "element", Label: "Use element",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<defs>
					<rect id="myRect" width="30" height="20" fill="#e74c3c"/>
					<circle id="myCircle" r="10" fill="#3498db"/>
				</defs>
				<use href="#myRect" x="10" y="10"/>
				<use href="#myRect" x="50" y="10"/>
				<use href="#myCircle" x="25" y="60"/>
				<use href="#myCircle" x="65" y="60"/>
			</svg>`,
		},

		// ── ClipPath ──

		{
			ID: "clip-basic", Category: "element", Label: "ClipPath",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<defs>
					<clipPath id="circleClip">
						<circle cx="50" cy="50" r="35"/>
					</clipPath>
				</defs>
				<rect x="10" y="10" width="80" height="80" fill="#e74c3c" clip-path="url(#circleClip)"/>
			</svg>`,
		},

		// ── Phase 6: Test Coverage ──

		{
			ID: "path-smooth-quad", Category: "path", Label: "Smooth quad T/t",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M10,50 Q30,10 50,50 T90,50" fill="none" stroke="#e74c3c" stroke-width="3"/>
			</svg>`,
		},
		{
			ID: "path-multi-subpath", Category: "path", Label: "Multi subpath",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<path d="M10,30 L40,10 L40,50 Z M60,30 L90,10 L90,50 Z M35,60 L65,60 L50,90 Z" fill="#3498db" stroke="#2c3e50" stroke-width="1"/>
			</svg>`,
		},
		{
			ID: "edge-degenerate", Category: "path", Label: "Degenerate cases",
			SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
				<rect x="10" y="10" width="0" height="0" fill="#e74c3c"/>
				<circle cx="30" cy="30" r="0" fill="#3498db"/>
				<rect x="10" y="50" width="30" height="20" rx="0" ry="0" fill="#2ecc71"/>
				<path d="M60,10 A0,0 0 1,1 60,10" fill="none" stroke="#f39c12" stroke-width="2"/>
				<path d="M60,50 L90,50 L90,80 L60,80 Z" fill="#9b59b6"/>
			</svg>`,
		},
	}
}
