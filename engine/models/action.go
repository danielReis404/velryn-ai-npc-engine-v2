package models

type ActionVerb string

const (
	ActionMove   ActionVerb = "MOVE"
	ActionTalk   ActionVerb = "TALK"
	ActionAttack ActionVerb = "ATTACK"
	ActionDefend ActionVerb = "DEFEND"
	ActionFlee   ActionVerb = "FLEE"
	ActionBuy    ActionVerb = "BUY"
	ActionSell   ActionVerb = "SELL"
	ActionTrade  ActionVerb = "TRADE"
	ActionUse    ActionVerb = "USE"
	ActionWait   ActionVerb = "WAIT"
)
