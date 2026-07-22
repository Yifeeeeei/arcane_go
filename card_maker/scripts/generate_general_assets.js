const path = require('path');
const sharp = require('sharp');

const outDir = path.resolve(__dirname, '../resources/general');

const defs = `
<defs>
  <linearGradient id="gold" x1="0" y1="0" x2="1" y2="1">
    <stop offset="0" stop-color="#7d4c16"/><stop offset="0.22" stop-color="#e8c36e"/>
    <stop offset="0.48" stop-color="#fff0a8"/><stop offset="0.72" stop-color="#bd7b25"/>
    <stop offset="1" stop-color="#6b3b10"/>
  </linearGradient>
  <filter id="soft" x="-20%" y="-20%" width="140%" height="140%">
    <feGaussianBlur in="SourceAlpha" stdDeviation="2" result="b"/>
    <feOffset in="b" dx="0" dy="2" result="o"/>
    <feColorMatrix in="o" type="matrix" values="0 0 0 0 0.13  0 0 0 0 0.07  0 0 0 0 0.02  0 0 0 .55 0"/>
    <feMerge><feMergeNode/><feMergeNode in="SourceGraphic"/></feMerge>
  </filter>
</defs>`;

function svg(w, h, body, extraDefs = '') {
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}" viewBox="0 0 ${w} ${h}">${defs}${extraDefs}${body}</svg>`;
}

const borders = {
  border_unit: svg(1180, 1720, `
    <g fill="none" stroke="url(#gold)" stroke-linejoin="round" stroke-linecap="round" filter="url(#soft)">
      <rect x="44" y="44" width="1092" height="1632" rx="10" stroke-width="22"/>
      <rect x="66" y="66" width="1048" height="1588" rx="6" stroke-width="6" opacity=".82"/>
      <path d="M70 128V70h58M1052 70h58v58M70 1592v58h58M1052 1650h58v-58" stroke-width="5" opacity=".8"/>
      <path d="M82 82l28 28M1098 82l-28 28M82 1638l28-28M1098 1638l-28-28" stroke-width="3" opacity=".65"/>
    </g>`),
  border_ability: svg(1180, 1720, `
    <g fill="none" stroke="url(#gold)" stroke-linecap="round" filter="url(#soft)">
      <rect x="46" y="46" width="1088" height="1628" rx="0" stroke-width="20"/>
      <rect x="66" y="66" width="1048" height="1588" rx="13" stroke-width="6" opacity=".8"/>
      <path d="M70 112c23-2 40-19 42-42M1068 70c2 23 19 40 42 42M70 1608c23 2 40 19 42 42M1068 1650c2-23 19-40 42-42" stroke-width="4" opacity=".72"/>
    </g>`),
  border_item: svg(1180, 1720, `
    <g fill="none" stroke="url(#gold)" stroke-linejoin="bevel" filter="url(#soft)">
      <rect x="44" y="44" width="1092" height="1632" stroke-width="22"/>
      <path d="M82 66h1016l16 16v1556l-16 16H82l-16-16V82z" stroke-width="6" opacity=".82"/>
      <path d="M70 122V70h52M1058 70h52v52M70 1598v52h52M1058 1650h52v-52" stroke-width="5" opacity=".72"/>
      <path d="M82 96h14V82M1084 82v14h14M82 1624h14v14M1084 1638v-14h14" stroke-width="3" opacity=".55"/>
    </g>`),
};

const ink = '#222833';
function circleIcon(color, body) {
  return svg(300, 300, `<g fill="${color}" stroke="${color}" stroke-linecap="round" stroke-linejoin="round">
    <circle cx="150" cy="150" r="137" fill="none" stroke-width="26"/>
    ${body}
  </g>`);
}

function fistIcon() {
  const color = '#c5681b';
  return circleIcon(color, `
    <defs>
      <mask id="frontFist">
        <rect width="300" height="300" fill="black"/>
        <g fill="white" stroke="white" stroke-linejoin="round">
          <path d="M105 190h90l-5 49h-80z" stroke="none"/>
          <rect x="84" y="118" width="132" height="91" rx="38" stroke="none"/>
          <rect x="88" y="91" width="32" height="72" rx="16" stroke="none"/>
          <rect x="121" y="80" width="32" height="83" rx="16" stroke="none"/>
          <rect x="154" y="84" width="32" height="79" rx="16" stroke="none"/>
          <rect x="187" y="96" width="30" height="67" rx="15" stroke="none"/>
          <path d="M99 133c-19 1-31 13-30 30 1 18 17 29 39 38l48 21c14 6 28 0 33-13 4-12-2-25-16-31l-43-18 42-3c13-1 22-11 21-23-1-12-11-20-24-19z" stroke="none"/>
        </g>
        <g fill="none" stroke="black" stroke-width="7" stroke-linecap="round">
          <path d="M120 97v29M153 88v38M186 101v27"/>
          <path d="M94 132c34 12 78 12 112 0" stroke-width="8"/>
          <path d="M119 158l52 22" stroke-width="8"/>
        </g>
      </mask>
    </defs>
    <g transform="translate(-27 -27) scale(1.18)">
      <rect x="60" y="48" width="180" height="202" mask="url(#frontFist)" stroke="none"/>
    </g>`);
}

const icons = {
  attack: fistIcon(),
  duration: circleIcon('#315b9f', `
    <circle cx="150" cy="150" r="87" fill="none" stroke-width="18"/>
    <path d="M150 150V91M150 150l45 30" fill="none" stroke-width="19"/>
    <circle cx="150" cy="150" r="12" stroke="none"/>`),
  ability_logo: svg(200, 200, `
    <path d="M100 18l18 55 55 27-55 27-18 55-18-55-55-27 55-27z" fill="${ink}"/>
    <path d="M151 31l7 20 20 7-20 7-7 20-7-20-20-7 20-7z" fill="${ink}"/>`),
  hero_logo: svg(200, 200, `
    <path d="M57 137L79 51l22-35 18 41 34 80c22 5 35 15 35 25 0 15-39 27-88 27s-88-12-88-27c0-11 17-20 45-25z" fill="#f7f4ed"/>
    <path d="M74 119c28 9 57 9 86-1l7 18c-39 13-79 13-118 1z" fill="#f7f4ed"/>`),
  item_logo: svg(200, 200, `
    <path d="M100 22l75 35c9 4 9 11 0 15l-75 30-75-30c-9-4-9-11 0-15z" fill="${ink}"/>
    <path d="M20 85l69 28v72L35 162c-10-4-15-12-15-23z" fill="${ink}"/>
    <path d="M111 113l69-28v54c0 11-5 19-15 23l-54 23z" fill="${ink}"/>
    <path d="M48 92v31c0 8 6 14 14 14s14-6 14-14v-19z" fill="none"/>`),
  life: circleIcon('#d92d3a', `
    <path d="M150 248c-36-32-94-76-94-133 0-35 22-59 51-59 21 0 36 12 43 33 7-21 22-33 43-33 29 0 51 24 51 59 0 57-58 101-94 133z" stroke="none"/>`),
  power: circleIcon('#7135b5', `
    <path d="M150 43l27 79 80 28-80 28-27 79-27-79-80-28 80-28z" stroke="none"/>`),
  unit_logo: svg(200, 200, `
    <g fill="${ink}">
      <circle cx="100" cy="68" r="34"/>
      <path d="M26 178c5-55 33-86 74-86s69 31 74 86c-39 18-109 18-148 0z"/>
    </g>`),
};

async function render(name, data) {
  const target = path.join(outDir, `${name}.png`);
  await sharp(Buffer.from(data)).png({ compressionLevel: 9 }).toFile(target);
}

(async () => {
  for (const [name, data] of Object.entries(borders)) await render(name, data);
  for (const [name, data] of Object.entries(icons)) await render(name, data);
  console.log(`Generated ${Object.keys(borders).length} borders and ${Object.keys(icons).length} icons.`);
})().catch(err => { console.error(err); process.exit(1); });
