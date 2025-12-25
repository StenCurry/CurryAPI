import apiClient from './client'

// ============================================================================
// TypeScript Interfaces for Request/Response Types
// ============================================================================

// Game Balance Response
export interface GameBalance {
  balance: number
  total_won: number
  total_lost: number
  total_exchanged: number
  games_played: number
  created_at?: string
  updated_at?: string
}

// Game Coin Transaction
export interface GameCoinTransaction {
  id: number
  type: string
  game_type?: string
  amount: number
  balance_after: number
  description?: string
  created_at: string
}

// Deduct Game Coins Request
export interface DeductGameCoinsRequest {
  amount: number
  game_type: string
  description?: string
}

// Add Game Coins Request
export interface AddGameCoinsRequest {
  amount: number
  game_type: string
  description?: string
}

// Game Coin Operation Response (for deduct/add)
export interface GameCoinOperationResponse {
  success: boolean
  transaction: GameCoinTransaction
  balance_after: number
}

// Reset Game Coins Response
export interface ResetGameCoinsResponse {
  success: boolean
  balance: number
  total_won: number
  total_lost: number
  total_exchanged: number
  games_played: number
}

// Game Transactions Response (paginated)
export interface GameTransactionsResponse {
  transactions: GameCoinTransaction[]
  total: number
  limit: number
  offset: number
}

// Migrate LocalStorage Request
export interface MigrateLocalStorageRequest {
  balance: number
  total_won: number
  total_lost: number
  games_played: number
}

// Migrate LocalStorage Response
export interface MigrateLocalStorageResponse {
  success: boolean
  balance: number
  total_won: number
  total_lost: number
  total_exchanged: number
  games_played: number
  migrated: boolean
}

// Exchange Record
export interface ExchangeRecord {
  id: number
  game_coins_amount: number
  usd_amount: number
  exchange_rate: number
  status: string
  created_at: string
}

// Exchange Game Coins Request
export interface ExchangeGameCoinsRequest {
  amount: number
}

// Exchange Game Coins Response
export interface ExchangeGameCoinsResponse {
  success: boolean
  exchange_record: ExchangeRecord
  new_game_balance: number
  new_account_balance: number
}

// Exchange History Response (paginated)
export interface ExchangeHistoryResponse {
  records: ExchangeRecord[]
  total: number
  limit: number
  offset: number
}

// Today's Exchange Amount Response
export interface TodayExchangeAmountResponse {
  amount: number
  limit: number
  remaining: number
}

// ============================================================================
// Game Coin API Functions
// ============================================================================

/**
 * Get current user's game coin balance
 * GET /api/game/balance
 * Requirements: 1.3, 7.3
 */
export const getGameBalance = () =>
  apiClient.get<GameBalance>('/api/game/balance')

/**
 * Deduct game coins from user's balance (for betting)
 * POST /api/game/deduct
 * Requirements: 1.2, 7.1
 */
export const deductGameCoins = (data: DeductGameCoinsRequest) =>
  apiClient.post<GameCoinOperationResponse>('/api/game/deduct', data)

/**
 * Add game coins to user's balance (for winning)
 * POST /api/game/add
 * Requirements: 1.2, 7.2
 */
export const addGameCoins = (data: AddGameCoinsRequest) =>
  apiClient.post<GameCoinOperationResponse>('/api/game/add', data)

/**
 * Reset user's game coin balance to initial value (100)
 * POST /api/game/reset
 * Requirements: 8.2
 */
export const resetGameCoins = () =>
  apiClient.post<ResetGameCoinsResponse>('/api/game/reset')

/**
 * Get paginated game coin transaction history
 * GET /api/game/transactions
 * Requirements: 1.6
 */
export const getGameTransactions = (limit = 20, offset = 0) =>
  apiClient.get<GameTransactionsResponse>('/api/game/transactions', {
    params: { limit, offset }
  })

/**
 * Migrate localStorage game data to database
 * POST /api/game/migrate
 * Requirements: 1.5
 */
export const migrateLocalStorage = (data: MigrateLocalStorageRequest) =>
  apiClient.post<MigrateLocalStorageResponse>('/api/game/migrate', data)

// ============================================================================
// Exchange API Functions
// ============================================================================

/**
 * Exchange game coins for account balance (USD)
 * POST /api/game/exchange
 * Requirements: 2.1, 2.2, 2.3, 2.6, 2.7
 */
export const exchangeGameCoins = (amount: number) =>
  apiClient.post<ExchangeGameCoinsResponse>('/api/game/exchange', { amount })

// Purchase Game Coins Response
export interface PurchaseGameCoinsResponse {
  success: boolean
  purchase_record: ExchangeRecord
  new_game_balance: number
  new_account_balance: number
}

/**
 * Purchase game coins with account balance (USD)
 * POST /api/game/purchase
 */
export const purchaseGameCoins = (amount: number) =>
  apiClient.post<PurchaseGameCoinsResponse>('/api/game/purchase', { amount })

/**
 * Get paginated exchange history for current user
 * GET /api/game/exchange/history
 * Requirements: 3.1, 3.2, 3.3
 */
export const getExchangeHistory = (limit = 20, offset = 0) =>
  apiClient.get<ExchangeHistoryResponse>('/api/game/exchange/history', {
    params: { limit, offset }
  })

/**
 * Get today's exchange amount and remaining limit
 * GET /api/game/exchange/today
 * Requirements: 2.7
 */
export const getTodayExchangeAmount = () =>
  apiClient.get<TodayExchangeAmountResponse>('/api/game/exchange/today')


// ============================================================================
// Admin Exchange API Functions
// ============================================================================

// Admin Exchange Record with User Info
export interface AdminExchangeRecord {
  id: number
  user_id: number
  username: string
  email: string
  game_coins_amount: number
  usd_amount: number
  exchange_rate: number
  status: string
  created_at: string
}

// Admin Exchange Records Response (paginated)
export interface AdminExchangeRecordsResponse {
  records: AdminExchangeRecord[]
  total: number
  limit: number
  offset: number
}

// Admin Exchange Statistics Response
export interface AdminExchangeStatsResponse {
  total_count: number
  total_usd: number
}

// Admin Exchange Query Parameters
export interface AdminExchangeQueryParams {
  user_id?: number
  start_date?: string
  end_date?: string
  limit?: number
  offset?: number
}

// ============================================================================
// Game Record Interfaces (Requirements: 1.1, 2.1, 3.3)
// ============================================================================

// Game-specific details interfaces
export interface WheelDetails {
  segment: string
  multiplier: number
}

export interface CoinDetails {
  choice: 'heads' | 'tails'
  coin_result: 'heads' | 'tails'
}

export interface NumberDetails {
  choice: 'big' | 'small'
  target: number
  actual_side: 'big' | 'small' | 'mid'
  is_exact_mid: boolean
}

// Union type for game details
export type GameDetails = WheelDetails | CoinDetails | NumberDetails

// Game Record
export interface GameRecord {
  id: number
  game_type: 'wheel' | 'coin' | 'number'
  bet_amount: number
  result: 'win' | 'lose'
  payout: number
  net_profit: number
  details: GameDetails
  created_at: string
}

// Create Game Record Request
export interface CreateGameRecordRequest {
  game_type: 'wheel' | 'coin' | 'number'
  bet_amount: number
  result: 'win' | 'lose'
  payout: number
  details: GameDetails
}

// Create Game Record Response
export interface CreateGameRecordResponse {
  success: boolean
  record: GameRecord
  stats: {
    games_played: number
    wins: number
    win_rate: string
    net_profit: string
  }
}

// Game Records Response (paginated)
export interface GameRecordsResponse {
  records: GameRecord[]
  total: number
  limit: number
  offset: number
}

// Game Stats Response
export interface GameStatsResponse {
  games_played: number
  wins: number
  losses: number
  win_rate: string
  net_profit: string
  total_won: string
  total_lost: string
}

// Leaderboard Entry
export interface LeaderboardEntry {
  rank: number
  user_id: number
  username: string
  total_winnings: number
  games_played: number
}

// Leaderboard Response
export interface LeaderboardResponse {
  entries: LeaderboardEntry[]
  current_user: LeaderboardEntry | null
  total_players: number
}

/**
 * Admin: Get all exchange records with optional filters
 * GET /admin/exchanges
 * Requirements: 6.1, 6.2, 6.3, 6.4
 */
export const getAdminExchangeRecords = (params?: AdminExchangeQueryParams) =>
  apiClient.get<AdminExchangeRecordsResponse>('/admin/exchanges', { params })

/**
 * Admin: Get exchange statistics
 * GET /admin/exchanges/stats
 * Requirements: 6.5
 */
export const getAdminExchangeStats = () =>
  apiClient.get<AdminExchangeStatsResponse>('/admin/exchanges/stats')

// ============================================================================
// Game Record API Functions
// ============================================================================

/**
 * Create a game record after completing a game
 * POST /api/game/record
 * Requirements: 1.1, 1.2, 1.3, 1.4
 */
export const createGameRecord = (data: CreateGameRecordRequest) =>
  apiClient.post<CreateGameRecordResponse>('/api/game/record', data)

/**
 * Get paginated game records for current user
 * GET /api/game/records
 * Requirements: 1.5, 1.6
 */
export const getGameRecords = (limit = 10, offset = 0) =>
  apiClient.get<GameRecordsResponse>('/api/game/records', {
    params: { limit, offset }
  })

/**
 * Get game statistics for current user
 * GET /api/game/stats
 * Requirements: 2.1
 */
export const getGameStats = () =>
  apiClient.get<GameStatsResponse>('/api/game/stats')

/**
 * Get global leaderboard
 * GET /api/game/leaderboard
 * Requirements: 3.1, 3.2
 */
export const getLeaderboard = (sort: 'winnings' | 'games' = 'winnings', limit = 10) =>
  apiClient.get<LeaderboardResponse>('/api/game/leaderboard', {
    params: { sort, limit }
  })
