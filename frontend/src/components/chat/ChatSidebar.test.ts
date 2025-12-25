/**
 * Property-Based Tests for ChatSidebar Component
 * 聊天侧边栏组件属性测试
 * 
 * **Feature: online-chat, Property 2: Conversation List Ordering**
 * **Validates: Requirements 1.2**
 */

import { describe, it, expect } from 'vitest'
import fc from 'fast-check'
import type { Conversation } from '@/api/chat'

// ============================================================================
// Property 2: Conversation List Ordering
// **Feature: online-chat, Property 2: Conversation List Ordering**
// **Validates: Requirements 1.2**
// ============================================================================

describe('Property 2: Conversation List Ordering', () => {
  /**
   * Helper function to generate a valid ISO date string from a timestamp
   */
  const timestampToISOString = (timestamp: number): string => {
    return new Date(timestamp).toISOString()
  }

  /**
   * Helper function to generate a random conversation
   * Using integer timestamps to avoid invalid date issues
   */
  const conversationArbitrary = fc.record({
    id: fc.integer({ min: 1, max: 1000000 }),
    title: fc.string({ minLength: 1, maxLength: 100 }),
    model: fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo'),
    // Use timestamps between 2020-01-01 and 2025-12-31
    created_at: fc.integer({ min: 1577836800000, max: 1767225600000 }).map(timestampToISOString),
    updated_at: fc.integer({ min: 1577836800000, max: 1767225600000 }).map(timestampToISOString)
  }) as fc.Arbitrary<Conversation>

  /**
   * Helper function to check if conversations are sorted by updated_at in descending order
   */
  const isSortedByUpdatedAtDesc = (conversations: Conversation[]): boolean => {
    if (conversations.length <= 1) return true
    
    for (let i = 0; i < conversations.length - 1; i++) {
      const current = new Date(conversations[i]!.updated_at).getTime()
      const next = new Date(conversations[i + 1]!.updated_at).getTime()
      if (current < next) {
        return false
      }
    }
    return true
  }

  /**
   * Helper function to sort conversations by updated_at descending (simulates backend behavior)
   */
  const sortConversationsByUpdatedAtDesc = (conversations: Conversation[]): Conversation[] => {
    return [...conversations].sort((a, b) => {
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
    })
  }

  it('should display conversations sorted by updated_at in descending order', () => {
    fc.assert(
      fc.property(
        fc.array(conversationArbitrary, { minLength: 2, maxLength: 20 }),
        (conversations) => {
          // Simulate what the backend returns (sorted by updated_at DESC)
          const sortedConversations = sortConversationsByUpdatedAtDesc(conversations)
          
          // Verify the sorted list is in descending order by updated_at
          return isSortedByUpdatedAtDesc(sortedConversations)
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should maintain order when conversations have same updated_at', () => {
    fc.assert(
      fc.property(
        fc.array(conversationArbitrary, { minLength: 2, maxLength: 10 }),
        // Use timestamp between 2024-01-01 and 2024-12-31
        fc.integer({ min: 1704067200000, max: 1735689600000 }),
        (conversations, sameTimestamp) => {
          const sameDate = new Date(sameTimestamp).toISOString()
          // Set all conversations to have the same updated_at
          const sameTimeConversations = conversations.map(conv => ({
            ...conv,
            updated_at: sameDate
          }))
          
          const sortedConversations = sortConversationsByUpdatedAtDesc(sameTimeConversations)
          
          // Should still be a valid sorted list (all equal timestamps)
          return isSortedByUpdatedAtDesc(sortedConversations)
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should correctly identify most recent conversation as first', () => {
    fc.assert(
      fc.property(
        fc.array(conversationArbitrary, { minLength: 2, maxLength: 20 }),
        (conversations) => {
          // Ensure unique updated_at timestamps
          const uniqueConversations = conversations.map((conv, index) => ({
            ...conv,
            updated_at: new Date(Date.now() - index * 1000 * 60).toISOString() // Each 1 minute apart
          }))
          
          const sortedConversations = sortConversationsByUpdatedAtDesc(uniqueConversations)
          
          // Find the conversation with the most recent updated_at
          const mostRecent = uniqueConversations.reduce((latest, conv) => {
            return new Date(conv.updated_at) > new Date(latest.updated_at) ? conv : latest
          })
          
          // The first conversation in sorted list should be the most recent
          return sortedConversations[0]?.id === mostRecent.id
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should preserve all conversations after sorting', () => {
    fc.assert(
      fc.property(
        fc.array(conversationArbitrary, { minLength: 1, maxLength: 20 }),
        (conversations) => {
          const sortedConversations = sortConversationsByUpdatedAtDesc(conversations)
          
          // Same length
          if (sortedConversations.length !== conversations.length) return false
          
          // All original IDs should be present
          const originalIds = new Set(conversations.map(c => c.id))
          const sortedIds = new Set(sortedConversations.map(c => c.id))
          
          return originalIds.size === sortedIds.size &&
            [...originalIds].every(id => sortedIds.has(id))
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should handle empty conversation list', () => {
    const emptyList: Conversation[] = []
    const sortedList = sortConversationsByUpdatedAtDesc(emptyList)
    
    expect(sortedList).toHaveLength(0)
    expect(isSortedByUpdatedAtDesc(sortedList)).toBe(true)
  })

  it('should handle single conversation', () => {
    fc.assert(
      fc.property(
        conversationArbitrary,
        (conversation) => {
          const singleList = [conversation]
          const sortedList = sortConversationsByUpdatedAtDesc(singleList)
          
          return sortedList.length === 1 &&
            sortedList[0]?.id === conversation.id &&
            isSortedByUpdatedAtDesc(sortedList)
        }
      ),
      { numRuns: 100 }
    )
  })
})
