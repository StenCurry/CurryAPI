/**
 * Markdown Rendering Utility
 * Markdown 渲染工具
 * 
 * Provides markdown rendering with syntax highlighting for code blocks.
 * Uses marked for markdown parsing and highlight.js for code highlighting.
 * 
 * Requirements: 4.1, 4.2, 4.3
 */

import { marked } from 'marked'
import hljs from 'highlight.js'

// Import highlight.js dark theme
import 'highlight.js/styles/vs2015.css'

// ============================================================================
// Configure marked with highlight.js
// ============================================================================

// Custom renderer for code blocks with syntax highlighting
const renderer = new marked.Renderer()

renderer.code = function({ text, lang }: { text: string; lang?: string }) {
  const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
  const highlighted = hljs.highlight(text, { language }).value
  return `<pre><code class="hljs language-${language}">${highlighted}</code></pre>`
}

// Configure marked options
marked.setOptions({
  renderer,
  gfm: true,        // GitHub Flavored Markdown
  breaks: true,     // Convert \n to <br>
})

// ============================================================================
// Export Functions
// ============================================================================

/**
 * Render markdown content to HTML
 * 将 Markdown 内容渲染为 HTML
 * 
 * @param content - Markdown content to render
 * @returns Rendered HTML string
 */
export function renderMarkdown(content: string): string {
  if (!content) return ''
  
  try {
    // marked.parse can return string or Promise<string>
    // We use synchronous mode here
    const result = marked.parse(content)
    return typeof result === 'string' ? result : ''
  } catch (error) {
    console.error('Markdown rendering error:', error)
    // Fallback: escape HTML and convert newlines
    return escapeHtml(content).replace(/\n/g, '<br>')
  }
}

/**
 * Escape HTML special characters
 * 转义 HTML 特殊字符
 * 
 * @param text - Text to escape
 * @returns Escaped text
 */
export function escapeHtml(text: string): string {
  const htmlEntities: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  }
  return text.replace(/[&<>"']/g, char => htmlEntities[char] || char)
}

/**
 * Check if content contains markdown syntax
 * 检查内容是否包含 Markdown 语法
 * 
 * @param content - Content to check
 * @returns True if content appears to contain markdown
 */
export function hasMarkdownSyntax(content: string): boolean {
  // Common markdown patterns
  const markdownPatterns = [
    /^#{1,6}\s/m,           // Headers
    /\*\*[^*]+\*\*/,        // Bold
    /\*[^*]+\*/,            // Italic
    /`[^`]+`/,              // Inline code
    /```[\s\S]*?```/,       // Code blocks
    /^\s*[-*+]\s/m,         // Unordered lists
    /^\s*\d+\.\s/m,         // Ordered lists
    /\[.+\]\(.+\)/,         // Links
    /^\s*>/m,               // Blockquotes
    /\|.+\|/,               // Tables
  ]
  
  return markdownPatterns.some(pattern => pattern.test(content))
}

export default {
  renderMarkdown,
  escapeHtml,
  hasMarkdownSyntax
}
