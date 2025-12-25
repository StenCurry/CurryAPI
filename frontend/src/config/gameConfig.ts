/**
 * Game Configuration Constants
 * 游戏配置常量
 * 
 * Requirements: 5.1, 6.1, 7.1
 */

import type { WheelSegment } from '@/utils/gameUtils'

/**
 * Initial game coins for new users
 * 新用户初始游戏币数量
 */
export const INITIAL_GAME_COINS = 100

/**
 * Lucky Wheel Game Configuration
 * 幸运转盘游戏配置
 * 
 * Requirements: 5.1
 */
export const wheelConfig = {
  minBet: 1,
  maxBet: 50,
  segments: [
    { label: '0x', multiplier: 0, color: '#6b7280', weight: 15 },
    { label: '0.5x', multiplier: 0.5, color: '#ef4444', weight: 20 },
    { label: '1x', multiplier: 1, color: '#f59e0b', weight: 25 },
    { label: '1.5x', multiplier: 1.5, color: '#10b981', weight: 18 },
    { label: '2x', multiplier: 2, color: '#3b82f6', weight: 12 },
    { label: '3x', multiplier: 3, color: '#8b5cf6', weight: 7 },
    { label: '5x', multiplier: 5, color: '#ec4899', weight: 3 },
  ] as WheelSegment[],
}

/**
 * Number Guess Game Configuration
 * 猜数字游戏配置 - 猜大小模式
 * 
 * Requirements: 6.1
 */
export const numberGuessConfig = {
  minBet: 1,
  maxBet: 50,
  range: { min: 1, max: 100 },
  // 猜大小模式：猜对翻倍
  payoutMultiplier: 1.9,
  // 中间值（1-100的中间是50）
  midPoint: 50,
}

/**
 * Coin Flip Game Configuration
 * 硬币翻转游戏配置
 * 
 * Requirements: 7.1
 */
export const coinFlipConfig = {
  minBet: 1,
  maxBet: 100,
  payoutMultiplier: 1.95,
}
