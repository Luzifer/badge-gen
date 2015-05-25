<svg xmlns="http://www.w3.org/2000/svg" width="{{ .Width }}" height="20">
   <linearGradient id="b" x2="0" y2="100%">
      <stop offset="0" stop-color="#bbb" stop-opacity=".1" />
      <stop offset="1" stop-opacity=".1" />
   </linearGradient>
   <mask id="a">
      <rect width="{{ .Width }}" height="20" rx="3" fill="#fff" />
   </mask>
   <g mask="url(#a)">
      <path fill="#555"    d="M0  0 h{{ .TitleWidth }}  v20 H0  z" />
      <path fill="#{{ .Color }}"    d="M{{ .TitleWidth }} 0 H{{ .Width }}  v20 H{{ .TitleWidth }} z" />
      <path fill="url(#b)" d="M0  0 h{{ .Width }} v20 H0  z" />
   </g>
   <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
      <text x="{{ .TitleAnchor }}" y="15" fill="#010101" fill-opacity=".3">{{ .Title }}</text>
      <text x="{{ .TitleAnchor }}" y="14">{{ .Title }}</text>
      <text x="{{ .TextAnchor }}" y="15" fill="#010101" fill-opacity=".3">{{ .Text }}</text>
      <text x="{{ .TextAnchor }}" y="14">{{ .Text }}</text>
   </g>
</svg>
