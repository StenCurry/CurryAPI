/**
 * Property-Based Tests for Markdown Rendering
 * Markdown 渲染属性测试
 * 
 * **Feature: online-chat, Property 11: Markdown Rendering Correctness**
 * **Validates: Requirements 4.1**
 */

import { describe, it, expect } from 'vitest'
import fc from 'fast-check'
import { renderMarkdown, escapeHtml, hasMarkdownSyntax } from './markdown'

// ============================================================================
// Property 11: Markdown Rendering Correctness
// **Feature: online-chat, Property 11: Markdown Rendering Correctness**
// **Validates: Requirements 4.1**
// ============================================================================

describe('Property 11: Markdown Rendering Correctness', () => {
  /**
   * Helper to check if HTML contains expected element
   */
  const containsElement = (html: string, tag: string): boolean => {
    const regex = new RegExp(`<${tag}[^>]*>`, 'i')
    return regex.test(html)
  }

  /**
   * Helper to check if HTML contains expected text content
   * Accounts for HTML entity encoding
   */
  const containsText = (html: string, text: string): boolean => {
    // Remove HTML tags and decode HTML entities for comparison
    const stripped = html.replace(/<[^>]*>/g, '')
    // Decode common HTML entities
    const decoded = stripped
      .replace(/&amp;/g, '&')
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&quot;/g, '"')
      .replace(/&#39;/g, "'")
    return decoded.includes(text)
  }
  
  /**
   * Safe characters for markdown text testing
   * Excludes HTML special chars and markdown syntax chars
   */
  const SAFE_CHARS = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  
  /**
   * Arbitrary for generating safe text without special characters
   * These characters get encoded or have special meaning in markdown
   */
  const safeTextArbitrary = (minLength: number, maxLength: number) =>
    fc.string({ minLength, maxLength }).map(s => 
      s.split('').filter(c => SAFE_CHARS.includes(c)).join('')
    ).filter(s => s.length >= minLength && s.trim().length > 0)

  // --------------------------------------------------------------------------
  // Header Rendering Tests
  // --------------------------------------------------------------------------

  it('should render headers correctly for all header levels', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 6 }),
        safeTextArbitrary(1, 50),
        (level, text) => {
          const headerPrefix = '#'.repeat(level)
          const markdown = `${headerPrefix} ${text}`
          const html = renderMarkdown(markdown)
          
          // Should contain the appropriate header tag
          return containsElement(html, `h${level}`) && containsText(html, text.trim())
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Bold Text Rendering Tests
  // --------------------------------------------------------------------------

  it('should render bold text with ** markers', () => {
    fc.assert(
      fc.property(
        safeTextArbitrary(1, 30).filter(s => !s.includes('*')),
        (text) => {
          const markdown = `**${text}**`
          const html = renderMarkdown(markdown)
          
          // Should contain strong tag and the text
          return containsElement(html, 'strong') && containsText(html, text)
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Italic Text Rendering Tests
  // --------------------------------------------------------------------------

  it('should render italic text with * markers', () => {
    fc.assert(
      fc.property(
        safeTextArbitrary(1, 30).filter(s => !s.includes('*') && !s.includes(' ')),
        (text) => {
          const markdown = `*${text}*`
          const html = renderMarkdown(markdown)
          
          // Should contain em tag and the text
          return containsElement(html, 'em') && containsText(html, text)
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Inline Code Rendering Tests
  // --------------------------------------------------------------------------

  it('should render inline code with backticks', () => {
    fc.assert(
      fc.property(
        safeTextArbitrary(1, 30).filter(s => !s.includes('`')),
        (code) => {
          const markdown = `\`${code}\``
          const html = renderMarkdown(markdown)
          
          // Should contain code tag and the code text
          return containsElement(html, 'code') && containsText(html, code)
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Code Block Rendering Tests
  // --------------------------------------------------------------------------

  it('should render code blocks with triple backticks', () => {
    fc.assert(
      fc.property(
        fc.constantFrom('javascript', 'typescript', 'python', 'go', 'plaintext'),
        fc.string({ minLength: 1, maxLength: 100 }).filter(s => 
          s.trim().length > 0 && 
          !s.includes('```')
        ),
        (language, code) => {
          const markdown = `\`\`\`${language}\n${code}\n\`\`\``
          const html = renderMarkdown(markdown)
          
          // Should contain pre and code tags
          return containsElement(html, 'pre') && containsElement(html, 'code')
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Unordered List Rendering Tests
  // --------------------------------------------------------------------------

  it('should render unordered lists correctly', () => {
    fc.assert(
      fc.property(
        fc.array(
          fc.string({ minLength: 1, maxLength: 30 }).filter(s => 
            s.trim().length > 0 && 
            !s.includes('\n') &&
            !s.startsWith('-') &&
            !s.startsWith('*')
          ),
          { minLength: 1, maxLength: 5 }
        ),
        (items) => {
          const markdown = items.map(item => `- ${item}`).join('\n')
          const html = renderMarkdown(markdown)
          
          // Should contain ul and li tags
          return containsElement(html, 'ul') && containsElement(html, 'li')
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Ordered List Rendering Tests
  // --------------------------------------------------------------------------

  it('should render ordered lists correctly', () => {
    fc.assert(
      fc.property(
        fc.array(
          fc.string({ minLength: 1, maxLength: 30 }).filter(s => 
            s.trim().length > 0 && 
            !s.includes('\n') &&
            !/^\d+\./.test(s)
          ),
          { minLength: 1, maxLength: 5 }
        ),
        (items) => {
          const markdown = items.map((item, i) => `${i + 1}. ${item}`).join('\n')
          const html = renderMarkdown(markdown)
          
          // Should contain ol and li tags
          return containsElement(html, 'ol') && containsElement(html, 'li')
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Link Rendering Tests
  // --------------------------------------------------------------------------

  it('should render links correctly', () => {
    fc.assert(
      fc.property(
        fc.string({ minLength: 1, maxLength: 20 }).filter(s => 
          s.trim().length > 0 && 
          !s.includes('[') && 
          !s.includes(']') &&
          !s.includes('\n')
        ),
        fc.webUrl(),
        (text, url) => {
          const markdown = `[${text}](${url})`
          const html = renderMarkdown(markdown)
          
          // Should contain anchor tag with href
          return containsElement(html, 'a') && html.includes('href=')
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Blockquote Rendering Tests
  // --------------------------------------------------------------------------

  it('should render blockquotes correctly', () => {
    fc.assert(
      fc.property(
        safeTextArbitrary(1, 50).filter(s => !s.startsWith('>')),
        (text) => {
          const markdown = `> ${text}`
          const html = renderMarkdown(markdown)
          
          // Should contain blockquote tag
          return containsElement(html, 'blockquote') && containsText(html, text)
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Content Preservation Tests
  // --------------------------------------------------------------------------

  it('should preserve text content after rendering', () => {
    fc.assert(
      fc.property(
        safeTextArbitrary(1, 100).filter(s => 
          // Exclude markdown special characters to test plain text
          !s.includes('#') &&
          !s.includes('*') &&
          !s.includes('`') &&
          !s.includes('[') &&
          !s.includes(']') &&
          !s.includes('-') &&
          !/^\d+\./.test(s)
        ),
        (text) => {
          const html = renderMarkdown(text)
          
          // The text content should be preserved in the output
          return containsText(html, text.trim())
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // Empty Input Tests
  // --------------------------------------------------------------------------

  it('should handle empty input gracefully', () => {
    expect(renderMarkdown('')).toBe('')
  })

  // --------------------------------------------------------------------------
  // escapeHtml Tests
  // --------------------------------------------------------------------------

  it('should escape HTML special characters', () => {
    fc.assert(
      fc.property(
        fc.string({ minLength: 0, maxLength: 100 }),
        (text) => {
          const escaped = escapeHtml(text)
          
          // Should not contain unescaped special characters
          // (unless they were already escaped in input)
          const hasUnescapedLt = escaped.includes('<') && !text.includes('<')
          const hasUnescapedGt = escaped.includes('>') && !text.includes('>')
          
          // 检查 & 是否被正确转义
          const ampEscapedCorrectly =
            !escaped.includes('&') ||
            escaped.includes('&amp;') ||
            escaped.includes('&lt;') ||
            escaped.includes('&gt;') ||
            escaped.includes('&quot;') ||
            escaped.includes('&#39;')

          return ampEscapedCorrectly && !hasUnescapedLt && !hasUnescapedGt
        }
      ),
      { numRuns: 100 }
    )
  })

  // --------------------------------------------------------------------------
  // hasMarkdownSyntax Tests
  // --------------------------------------------------------------------------

  it('should detect markdown syntax in content', () => {
    // Test known markdown patterns
    expect(hasMarkdownSyntax('# Header')).toBe(true)
    expect(hasMarkdownSyntax('**bold**')).toBe(true)
    expect(hasMarkdownSyntax('*italic*')).toBe(true)
    expect(hasMarkdownSyntax('`code`')).toBe(true)
    expect(hasMarkdownSyntax('```\ncode block\n```')).toBe(true)
    expect(hasMarkdownSyntax('- list item')).toBe(true)
    expect(hasMarkdownSyntax('1. ordered item')).toBe(true)
    expect(hasMarkdownSyntax('[link](url)')).toBe(true)
    expect(hasMarkdownSyntax('> quote')).toBe(true)
    
    // Plain text should not be detected as markdown
    expect(hasMarkdownSyntax('plain text')).toBe(false)
    expect(hasMarkdownSyntax('hello world')).toBe(false)
  })
})
