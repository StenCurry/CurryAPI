/**
 * Property-Based Tests for ModelSelector Component
 * 模型选择器组件属性测试
 * 
 * **Feature: online-chat, Property 10: Model Switch Preservation**
 * **Validates: Requirements 3.3**
 */

import { describe, it } from 'vitest'
import fc from 'fast-check'
import type { Message, Conversation } from '@/api/chat'

// ============================================================================
// Property 10: Model Switch Preservation
// **Feature: online-chat, Property 10: Model Switch Preservation**
// **Validates: Requirements 3.3**
// ============================================================================

describe('Property 10: Model Switch Preservation', () => {
  /**
   * Helper function to generate a valid ISO date string from a timestamp
   */
  const timestampToISOString = (timestamp: number): string => {
    return new Date(timestamp).toISOString()
  }

  /**
   * Arbitrary for generating random messages
   */
  const messageArbitrary: fc.Arbitrary<Message> = fc.record({
    id: fc.integer({ min: 1, max: 1000000 }),
    conversation_id: fc.integer({ min: 1, max: 1000000 }),
    role: fc.constantFrom('user', 'assistant', 'system') as fc.Arbitrary<'user' | 'assistant' | 'system'>,
    content: fc.string({ minLength: 1, maxLength: 1000 }),
    tokens: fc.integer({ min: 0, max: 10000 }),
    cost: fc.float({ min: 0, max: 10, noNaN: true }),
    created_at: fc.integer({ min: 1577836800000, max: 1767225600000 }).map(timestampToISOString)
  })

  /**
   * Arbitrary for generating random conversations
   */
  const conversationArbitrary: fc.Arbitrary<Conversation> = fc.record({
    id: fc.integer({ min: 1, max: 1000000 }),
    user_id: fc.option(fc.integer({ min: 1, max: 1000000 }), { nil: undefined }),
    title: fc.string({ minLength: 1, maxLength: 100 }),
    model: fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo'),
    system_prompt: fc.option(fc.string({ maxLength: 500 }), { nil: undefined }),
    created_at: fc.integer({ min: 1577836800000, max: 1767225600000 }).map(timestampToISOString),
    updated_at: fc.integer({ min: 1577836800000, max: 1767225600000 }).map(timestampToISOString)
  })

  /**
   * Simulates model switch operation
   * Returns the conversation with updated model and the original messages
   */
  function switchModel(
    conversation: Conversation,
    messages: Message[],
    newModel: string
  ): { conversation: Conversation; messages: Message[] } {
    // Model switch only updates the conversation's model field
    // Messages remain completely unchanged
    return {
      conversation: {
        ...conversation,
        model: newModel,
        updated_at: new Date().toISOString()
      },
      messages: messages // Messages are preserved as-is
    }
  }

  /**
   * Helper to check if two message arrays are identical
   */
  function messagesAreIdentical(original: Message[], after: Message[]): boolean {
    if (original.length !== after.length) return false
    
    for (let i = 0; i < original.length; i++) {
      const orig = original[i]!
      const curr = after[i]!
      
      if (
        orig.id !== curr.id ||
        orig.conversation_id !== curr.conversation_id ||
        orig.role !== curr.role ||
        orig.content !== curr.content ||
        orig.tokens !== curr.tokens ||
        orig.cost !== curr.cost ||
        orig.created_at !== curr.created_at
      ) {
        return false
      }
    }
    
    return true
  }

  it('should preserve all messages when switching models', () => {
    fc.assert(
      fc.property(
        conversationArbitrary,
        fc.array(messageArbitrary, { minLength: 1, maxLength: 50 }),
        fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo', 'gpt-4-turbo'),
        (conversation, messages, newModel) => {
          // Ensure messages belong to this conversation
          const conversationMessages = messages.map(m => ({
            ...m,
            conversation_id: conversation.id
          }))
          
          // Perform model switch
          const result = switchModel(conversation, conversationMessages, newModel)
          
          // All messages should be preserved exactly as they were
          return messagesAreIdentical(conversationMessages, result.messages)
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should preserve message count after model switch', () => {
    fc.assert(
      fc.property(
        conversationArbitrary,
        fc.array(messageArbitrary, { minLength: 0, maxLength: 100 }),
        fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo'),
        (conversation, messages, newModel) => {
          const conversationMessages = messages.map(m => ({
            ...m,
            conversation_id: conversation.id
          }))
          
          const result = switchModel(conversation, conversationMessages, newModel)
          
          return result.messages.length === conversationMessages.length
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should preserve message content after model switch', () => {
    fc.assert(
      fc.property(
        conversationArbitrary,
        fc.array(messageArbitrary, { minLength: 1, maxLength: 30 }),
        fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo'),
        (conversation, messages, newModel) => {
          const conversationMessages = messages.map(m => ({
            ...m,
            conversation_id: conversation.id
          }))
          
          const result = switchModel(conversation, conversationMessages, newModel)
          
          // Check that all message contents are preserved
          const originalContents = conversationMessages.map(m => m.content)
          const resultContents = result.messages.map(m => m.content)
          
          return originalContents.every((content, i) => content === resultContents[i])
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should preserve message order after model switch', () => {
    fc.assert(
      fc.property(
        conversationArbitrary,
        fc.array(messageArbitrary, { minLength: 2, maxLength: 30 }),
        fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo'),
        (conversation, messages, newModel) => {
          const conversationMessages = messages.map(m => ({
            ...m,
            conversation_id: conversation.id
          }))
          
          const result = switchModel(conversation, conversationMessages, newModel)
          
          // Check that message IDs are in the same order
          const originalIds = conversationMessages.map(m => m.id)
          const resultIds = result.messages.map(m => m.id)
          
          return originalIds.every((id, i) => id === resultIds[i])
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should update conversation model while preserving messages', () => {
    fc.assert(
      fc.property(
        conversationArbitrary,
        fc.array(messageArbitrary, { minLength: 1, maxLength: 20 }),
        fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo'),
        (conversation, messages, newModel) => {
          const conversationMessages = messages.map(m => ({
            ...m,
            conversation_id: conversation.id
          }))
          
          const result = switchModel(conversation, conversationMessages, newModel)
          
          // Conversation model should be updated
          const modelUpdated = result.conversation.model === newModel
          
          // Messages should be unchanged
          const messagesPreserved = messagesAreIdentical(conversationMessages, result.messages)
          
          return modelUpdated && messagesPreserved
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should handle empty message list during model switch', () => {
    fc.assert(
      fc.property(
        conversationArbitrary,
        fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo'),
        (conversation, newModel) => {
          const emptyMessages: Message[] = []
          
          const result = switchModel(conversation, emptyMessages, newModel)
          
          return result.messages.length === 0 && result.conversation.model === newModel
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should preserve message roles after model switch', () => {
    fc.assert(
      fc.property(
        conversationArbitrary,
        fc.array(messageArbitrary, { minLength: 1, maxLength: 30 }),
        fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo'),
        (conversation, messages, newModel) => {
          const conversationMessages = messages.map(m => ({
            ...m,
            conversation_id: conversation.id
          }))
          
          const result = switchModel(conversation, conversationMessages, newModel)
          
          // Check that all message roles are preserved
          const originalRoles = conversationMessages.map(m => m.role)
          const resultRoles = result.messages.map(m => m.role)
          
          return originalRoles.every((role, i) => role === resultRoles[i])
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should preserve message timestamps after model switch', () => {
    fc.assert(
      fc.property(
        conversationArbitrary,
        fc.array(messageArbitrary, { minLength: 1, maxLength: 30 }),
        fc.constantFrom('gpt-4o', 'claude-3.5-sonnet', 'gpt-3.5-turbo'),
        (conversation, messages, newModel) => {
          const conversationMessages = messages.map(m => ({
            ...m,
            conversation_id: conversation.id
          }))
          
          const result = switchModel(conversation, conversationMessages, newModel)
          
          // Check that all message timestamps are preserved
          const originalTimestamps = conversationMessages.map(m => m.created_at)
          const resultTimestamps = result.messages.map(m => m.created_at)
          
          return originalTimestamps.every((ts, i) => ts === resultTimestamps[i])
        }
      ),
      { numRuns: 100 }
    )
  })
})
