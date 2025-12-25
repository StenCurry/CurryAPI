/**
 * Game Utility Functions
 * 游戏工具函数模块
 * 
 * Requirements: 1.1, 2.2, 5.2, 5.4, 6.3, 6.5, 7.5
 */

export interface WheelSegment {
  label: string
  multiplier: number
  color: string
  weight?: number
}

export interface BetValidationResult {
  valid: boolean
  error?: string
}

/**
 * Generate a random integer between min and max (inclusive)
 * 生成指定范围内的随机整数
 * 
 * @param min - Minimum value (inclusive)
 * @param max - Maximum value (inclusive)
 * @returns Random integer in range [min, max]
 * 
 * Requirements: 6.3
 */
export function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

/**
 * Generate a random boolean value
 * 生成随机布尔值
 * 
 * @returns Random boolean (true or false)
 * 
 * Requirements: 7.5
 */
export function randomBoolean(): boolean {
  return Math.random() < 0.5
}

/**
 * Spin the wheel and return the index of the selected segment
 * 转动转盘并返回选中的扇区索引
 * 
 * Uses weighted random selection based on segment weights
 * 
 * @param segments - Array of wheel segments with optional weights
 * @returns Index of the selected segment
 * 
 * Requirements: 5.4
 */
export function spinWheel(segments: WheelSegment[]): number {
  if (segments.length === 0) return -1
  const totalWeight = segments.reduce((sum, s) => sum + (s.weight ?? 1), 0)
  let random = Math.random() * totalWeight
  for (let i = 0; i < segments.length; i++) {
    const segment = segments[i]!
    random -= segment.weight ?? 1
    if (random <= 0) return i
  }
  return segments.length - 1
}

/**
 * Validate a bet amount against game constraints
 * 验证下注金额是否符合游戏约束
 * 
 * @param betAmount - The amount to bet
 * @param gameCoins - User's current game coin balance
 * @param minBet - Minimum allowed bet
 * @param maxBet - Maximum allowed bet
 * @returns Validation result with valid flag and optional error message
 * 
 * Requirements: 5.2, 6.2, 7.2
 */
export function validateBet(
  betAmount: number,
  gameCoins: number,
  minBet: number,
  maxBet: number
): BetValidationResult {
  if (betAmount <= 0) return { valid: false, error: '下注金额必须大于0' }
  if (betAmount > gameCoins) return { valid: false, error: '游戏币不足' }
  if (betAmount < minBet) return { valid: false, error: `最低下注 ${minBet} 游戏币` }
  if (betAmount > maxBet) return { valid: false, error: `最高下注 ${maxBet} 游戏币` }
  return { valid: true }
}

/**
 * Calculate payout amount with 2 decimal precision
 * 计算赔付金额（保留2位小数）
 * 
 * @param betAmount - The original bet amount
 * @param multiplier - The payout multiplier
 * @returns Payout amount rounded to 2 decimal places
 * 
 * Requirements: 5.4, 6.5, 7.5
 */
export function calculatePayout(betAmount: number, multiplier: number): number {
  return Number((betAmount * multiplier).toFixed(2))
}

/**
 * Get time-based greeting message
 * 根据当前时间返回问候语
 * 
 * @returns Greeting string based on current hour
 * 
 * Requirements: 1.1
 */
export function getGreeting(): string {
  const hour = new Date().getHours()
  if (hour < 12) return '早上好'
  if (hour < 18) return '下午好'
  return '晚上好'
}

/**
 * Calculate account age in days
 * 计算账户年龄（天数）
 * 
 * @param createdAt - Account creation date string
 * @returns Number of days since account creation
 * 
 * Requirements: 2.2
 */
export function calculateAccountAge(createdAt: string): number {
  const created = new Date(createdAt)
  const now = new Date()
  return Math.floor((now.getTime() - created.getTime()) / (1000 * 60 * 60 * 24))
}
