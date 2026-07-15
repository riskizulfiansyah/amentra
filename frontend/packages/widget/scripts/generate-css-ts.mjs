import { readFileSync, writeFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const cssPath = resolve(__dirname, '../src/styles.css')
const tsPath = resolve(__dirname, '../src/styles.ts')

const css = readFileSync(cssPath, 'utf8')

const escaped = css
  .replace(/\\/g, '\\\\')
  .replace(/`/g, '\\`')
  .replace(/\$/g, '\\$')

const content = `const css = \`${escaped}\`\nexport default css\n`
writeFileSync(tsPath, content)

console.log('Generated src/styles.ts')
