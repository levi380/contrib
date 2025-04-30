package helper

// 账变定义
const (
	// 会员相关
	TransactionMemberClass = 1000 //会员大类
	TransactionCommission  = 1001 //佣金
	TransactionAdjust      = 1002 //调整

	// 财务相关
	TransactionFinanceClass   = 2000 //财务大类
	TransactionDeposit        = 2001 //存款
	TransactionWithdraw       = 2002 //取款
	TransactionWithdrawCancel = 2003 //取消取款

	// 场馆游戏相关
	TransactionGameClass            = 3000 //场馆大类
	TransactionBet                  = 3001 //投注
	TransactionBetCancel            = 3002 //投注取消
	TransactionPayout               = 3003 //派彩
	TransactionResettlePlus         = 3004 //重新结算加币
	TransactionResettleDeduction    = 3005 //重新结算减币
	TransactionCancelPayout         = 3006 //取消派彩
	TransactionSettledBetCancel     = 3007 //投注取消(已结算注单)
	TransactionCancelledBetRollback = 3008 //已取消注单回滚
	TransactionBetNSettleWin        = 3009 //老虎机投付赢
	TransactionBetNSettleLose       = 3010 //老虎机投付输
	TransactionAdjustPlus           = 3011 //场馆调整加
	TransactionVenueRebate          = 3013 //场馆返利
	TransactionAdjustDiv            = 3012 //场馆调整减
	TransactionEVOPrize             = 3014 //游戏奖金(EVO)
	TransactionEVOPromote           = 3015 //推广(EVO)
	TransactionEVOJackpot           = 3016 //头奖(EVO)
	TransactionBonusWin             = 3017 //免费旋转奖励(PP)
	TransactionJackPotWin           = 3018 //累积奖金赢奖(PP)
	TransactionPromoWin             = 3019 //促销赢奖(PP)
	TransactionMGTournament         = 3020 //擂台賽(MG) TOURNAMENT
	TransactionMGPromotion          = 3021 //促進活動(MG) PROMOTION
	TransactionMGAchievement        = 3022 //排行榜(MG) ACHIEVEMENT
	TransactionMGStore              = 3023 //商店(MG) STORE
	TransactionCQ9Payoff            = 3024 //活动派彩(CQ9)
	TransactionVenueTransferIn      = 3025 //转入场馆
	TransactionVenueTransferOut     = 3026 //转出场馆
	TransactionTip                  = 3027 //打赏
	TransactionCancelTip            = 3028 //取消打赏
	TransactionPromoDeduct          = 3029 //场馆活动扣款
	TransactionPromoPayout          = 3030 //场馆活动加款

	// 活动相关
	TransactionPromoClass         = 4000 //活动大类
	TransactionPromoBonusBlindBox = 4001 //盲盒活动彩金
	TransactionPromoBonusSuperAce = 4002 //SuperAce活动彩金
)
