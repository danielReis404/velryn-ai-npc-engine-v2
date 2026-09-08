package engine

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"ai-agent-engine/models"
)

func (w *World) resolveIntent(ctx context.Context, intent TurnIntent) {
	npc := intent.Actor
	action := intent.Action
	if npc.HP <= 0 {
		return
	}
	startPos := npc.Position

	if npc.IsLLM && (intent.Source == SourceAI || intent.Source == SourceAgentBatch) {
		if action.Action == models.ActionTalk && action.Target != "" {
			if action.Target == npc.NegotiationPartner {
				npc.NegotiationTicks++
			} else {
				npc.NegotiationPartner = action.Target
				npc.NegotiationTicks = 1
			}
		} else {
			npc.NegotiationPartner = ""
			npc.NegotiationTicks = 0
		}
	}

	var headerLine string
	collapsible := false
	switch intent.Source {
	case SourceAIBatch, SourceAgentBatch:
		headerLine = fmt.Sprintf("\n[%s via %s 🧠🧠 AI call (grupo)]: %s\n", npc.Name, orDefault(npc.LastProvider, "unknown"), action.InternalThought)
	case SourceInstinct:
		headerLine = fmt.Sprintf("\n[%s instinct, no AI call]: %s\n", npc.Name, action.InternalThought)
	case SourceAI:
		headerLine = fmt.Sprintf("\n[%s via %s 🧠 AI call]: %s\n", npc.Name, orDefault(npc.LastProvider, "unknown"), action.InternalThought)
	case SourceExternal:
		headerLine = fmt.Sprintf("\n[%s 🎮 player action]: %s\n", npc.Name, action.InternalThought)
	case SourcePlayerIdle:
		collapsible = true
		headerLine = fmt.Sprintf("\n[%s 🎮 waiting for player]\n", npc.Name)
	case SourceFallbackBatch, SourceFallbackBatchMissing:
		headerLine = fmt.Sprintf("\n[%s ⚠️ agent-service unavailable, local fallback]: %s\n", npc.Name, action.InternalThought)
	case SourceScripted:
		collapsible = true
		if npc.IsMerchant {
			headerLine = fmt.Sprintf("\n[%s the merchant]: tends the shop.\n", npc.Name)
		} else {
			headerLine = fmt.Sprintf("\n[%s (Lvl %d) scripted monster]: acts...\n", npc.Name, npc.Level)
		}
	default:
		collapsible = true
		headerLine = fmt.Sprintf("\n[%s routine, no AI call]: %s\n", npc.Name, action.InternalThought)
	}

	if collapsible && headerLine == npc.LastPrintedLine {

		npc.IdleRepeatCount++
	} else {
		if npc.IdleRepeatCount > 1 {
			fmt.Printf("    ...(%s repeated the above %d times)\n", npc.Name, npc.IdleRepeatCount)
		}
		fmt.Print(headerLine)
		npc.LastPrintedLine = headerLine
		npc.IdleRepeatCount = 1
	}

	var memoryEntry string

	switch action.Action {
	case models.ActionMove:
		if len(action.TargetCoordinates) == 2 {
			newX, newY := action.TargetCoordinates[0], action.TargetCoordinates[1]
			if newX >= 0 && newX < w.Width && newY >= 0 && newY < w.Height &&
				w.distance(npc.Position, models.Position{X: newX, Y: newY}) <= 1 &&
				!w.isTileOccupied(models.Position{X: newX, Y: newY}, npc.Name) {
				npc.Position.X, npc.Position.Y = newX, newY
				tile := w.Map[newX][newY]
				fmt.Printf("[%s] moved to [%d, %d] (%s)\n", npc.Name, newX, newY, tile.Biome)
				memoryEntry = fmt.Sprintf("I moved to [%d, %d], entering the %s.", newX, newY, tile.Biome)
				npc.LastAction = fmt.Sprintf("moved to %s.", tile.Biome)
			} else {
				memoryEntry = "I tried to move but failed."
				npc.LastAction = "stumbled."
			}
		}

	case models.ActionFlee:
		if len(action.TargetCoordinates) == 2 {
			newX, newY := action.TargetCoordinates[0], action.TargetCoordinates[1]
			if newX >= 0 && newX < w.Width && newY >= 0 && newY < w.Height &&
				w.distance(npc.Position, models.Position{X: newX, Y: newY}) <= 1 &&
				!w.isTileOccupied(models.Position{X: newX, Y: newY}, npc.Name) {
				npc.Position.X, npc.Position.Y = newX, newY
				tile := w.Map[newX][newY]
				fmt.Printf("[%s] retreats to [%d, %d] (%s)\n", npc.Name, newX, newY, tile.Biome)
				memoryEntry = fmt.Sprintf("I fled to [%d, %d] to escape danger.", newX, newY)
				npc.LastAction = "fled from danger."
			} else {
				memoryEntry = "I tried to flee but was cornered."
				npc.LastAction = "was cornered."
			}
		}

	case models.ActionAttack:
		targetName := action.Target
		targetNPC, exists := w.NPCs[targetName]

		currentTile := w.Map[npc.Position.X][npc.Position.Y]
		var targetTile *models.Tile
		if exists {
			targetTile = w.Map[targetNPC.Position.X][targetNPC.Position.Y]
		}

		if currentTile.IsSafeZone || (exists && targetTile.IsSafeZone) {
			fmt.Printf("[%s] tried to attack, but the Architects' barrier nullifies violence here!\n", npc.Name)
			memoryEntry = "I tried to fight, but the magic barrier of the city prevented it."
			npc.LastAction = "was stopped by a magical barrier."
			break
		}

		if exists && w.isAdjacent(npc, targetNPC) && targetNPC.HP > 0 {
			foeLabel := targetName
			if targetNPC.Species != "" {
				foeLabel = targetNPC.Species
			}

			skillName := action.Skill
			if skillName == "" {
				skillName = "a basic strike"
			}

			totalDamage := npc.BaseAttack
			weaponBroke := false
			if npc.Equipped != nil {
				totalDamage += npc.Equipped.Power
				npc.Equipped.Durability--
				fmt.Printf("[%s] swings their %s (Durability: %d/%d)\n", npc.Name, npc.Equipped.Name, npc.Equipped.Durability, npc.Equipped.MaxDurability)
				if npc.Equipped.Durability <= 0 {
					fmt.Printf("[%s]'s %s broke!\n", npc.Name, npc.Equipped.Name)
					npc.Equipped = nil
					weaponBroke = true
				}
			}

			if targetNPC.Defending {
				totalDamage /= 2
				targetNPC.Defending = false
				fmt.Printf("[%s] had braced for it — damage cut to %d!\n", targetNPC.Name, totalDamage)
			}
			targetNPC.HP -= totalDamage
			targetNPC.TicksSinceLastHit = 0

			if !npc.IsLLM && !npc.IsMerchant {

				npc.LastAttackTarget = targetName
			}

			if !targetNPC.IsLLM && !targetNPC.IsMerchant && targetNPC.LastAttackTarget != "" && targetNPC.LastAttackTarget != npc.Name {

				if rescued, ok := w.NPCs[targetNPC.LastAttackTarget]; ok && rescued.HP > 0 && rescued.IsLLM {
					adjustTrust(rescued, npc.Name, 15)
				}
			}

			fmt.Printf("[%s] uses %s on [%s] for %d damage! (%s has %d HP left)\n", npc.Name, skillName, targetName, totalDamage, targetName, targetNPC.HP)
			memoryEntry = fmt.Sprintf("I used %s against the %s and dealt %d damage.", skillName, foeLabel, totalDamage)
			if weaponBroke {
				memoryEntry += " My weapon broke in the process."
			}
			npc.LastAction = fmt.Sprintf("used %s against the %s.", skillName, foeLabel)

			if npc.IsLLM && targetNPC.IsLLM {
				adjustTrust(targetNPC, npc.Name, -25)
				adjustFear(targetNPC, npc.Name, 20)
				adjustTrust(npc, targetNPC.Name, -10)

				for _, bystander := range w.NPCs {
					if bystander.Name == npc.Name || bystander.Name == targetNPC.Name || !bystander.IsLLM {
						continue
					}
					if !w.isAdjacent(bystander, targetNPC) {
						continue
					}
					bAction, decided := w.currentTickActions[bystander.Name]
					helped := decided && bAction.Action == "ATTACK" && bAction.Target == npc.Name
					if !helped {
						adjustTrust(targetNPC, bystander.Name, -5)
					}
				}
			}

			if targetNPC.HP <= 0 {
				fmt.Printf("[%s] has DIED! [%s] is victorious!\n", targetName, npc.Name)
				memoryEntry = fmt.Sprintf("I defeated the %s using %s.", foeLabel, skillName)
				if weaponBroke {
					memoryEntry += " My weapon broke in the process."
				}
				xpGained := targetNPC.Level * 15
				npc.XP += xpGained
				if npc.XP >= (npc.Level * 20) {
					npc.Level++
					npc.MaxHP += 5
					npc.HP = npc.MaxHP
					npc.BaseAttack += 2
					fmt.Printf("LEVEL UP! [%s] is now Level %d!\n", npc.Name, npc.Level)
				}

				if t, ok := monsterTemplate(targetNPC.Species); ok {
					goldReward := t.GoldMin
					if t.GoldMax > t.GoldMin {
						goldReward += rand.Intn(t.GoldMax - t.GoldMin + 1)
					}
					npc.Gold += goldReward
					memoryEntry += fmt.Sprintf(" I looted %d gold.", goldReward)
					fmt.Printf("[%s] loots %d gold from the %s.\n", npc.Name, goldReward, foeLabel)

					if len(models.LootPool) > 0 && rand.Float64() < t.LootChance {
						drop := models.LootPool[rand.Intn(len(models.LootPool))]
						npc.Inventory = append(npc.Inventory, &drop)
						memoryEntry += fmt.Sprintf(" It also dropped a %s!", drop.Name)
						fmt.Printf("[%s] finds a %s on the %s's remains!\n", npc.Name, drop.Name, foeLabel)
					}
				}

				if targetNPC.IsLLM {

					w.reviveDownedNPC(ctx, targetNPC)
				} else {
					w.recentMonsterKills = append(w.recentMonsterKills, targetNPC.Species)
					if len(w.recentMonsterKills) > 10 {
						w.recentMonsterKills = w.recentMonsterKills[len(w.recentMonsterKills)-10:]
					}
					delete(w.NPCs, targetName)
				}
			}
		} else {
			notice := "wasn't close enough to hit"
			if exists && targetNPC.HP <= 0 {
				notice = "was already down"
			} else if !exists {
				notice = "wasn't there anymore"
			}

			fmt.Printf("[%s] attacked, but %s %s.\n", npc.Name, targetName, notice)
			memoryEntry = fmt.Sprintf("I tried to attack %s, but they %s — I wasn't actually close enough.", targetName, notice)
			npc.LastAction = fmt.Sprintf("tried to attack %s from too far away.", targetName)
		}

	case models.ActionDefend:
		npc.Defending = true
		fmt.Printf("[%s] braces for the next hit.\n", npc.Name)
		memoryEntry = "I braced myself, ready to absorb the next attack."
		npc.LastAction = "braced for an attack."

	case models.ActionBuy:
		merchant, ok := w.NPCs[action.Target]
		if !ok || !merchant.IsMerchant || !w.isAdjacent(npc, merchant) {
			memoryEntry = "I looked for a merchant to buy from, but none was close enough."
			npc.LastAction = "found no merchant nearby."
			break
		}
		item := findShopItem(merchant, action.Item)
		if item == nil {
			memoryEntry = fmt.Sprintf("I asked %s for a %s, but they don't stock that.", merchant.Name, action.Item)
			npc.LastAction = "asked about an item the merchant doesn't sell."
			break
		}
		if npc.Gold < item.Price {
			memoryEntry = fmt.Sprintf("I wanted a %s (%d gold), but I only had %d.", item.Name, item.Price, npc.Gold)
			npc.LastAction = "couldn't afford an item."
			break
		}
		npc.Gold -= item.Price
		bought := *item
		if bought.Kind == "weapon" {
			if npc.Equipped != nil {
				npc.Inventory = append(npc.Inventory, npc.Equipped)
			}
			npc.Equipped = &bought
			memoryEntry = fmt.Sprintf("I bought and equipped a %s from %s for %d gold.", bought.Name, merchant.Name, bought.Price)
		} else {
			npc.Inventory = append(npc.Inventory, &bought)
			memoryEntry = fmt.Sprintf("I bought a %s from %s for %d gold.", bought.Name, merchant.Name, bought.Price)
		}
		npc.LastAction = fmt.Sprintf("bought a %s.", bought.Name)
		fmt.Printf("[%s] buys a %s from %s for %d gold (%d gold left).\n", npc.Name, bought.Name, merchant.Name, bought.Price, npc.Gold)

	case models.ActionSell:
		merchant, ok := w.NPCs[action.Target]
		if !ok || !merchant.IsMerchant || !w.isAdjacent(npc, merchant) {
			memoryEntry = "I looked for a merchant to sell to, but none was close enough."
			npc.LastAction = "found no merchant nearby."
			break
		}
		idx := findInventoryIndex(npc, action.Item)
		if idx == -1 {
			memoryEntry = fmt.Sprintf("I tried to sell a %s, but I don't have one.", action.Item)
			npc.LastAction = "tried to sell something they didn't have."
			break
		}
		sold := npc.Inventory[idx]
		npc.Inventory = append(npc.Inventory[:idx], npc.Inventory[idx+1:]...)
		price := sold.Price / 2
		if price < 1 {
			price = 1
		}
		npc.Gold += price
		memoryEntry = fmt.Sprintf("I sold my %s to %s for %d gold.", sold.Name, merchant.Name, price)
		npc.LastAction = fmt.Sprintf("sold a %s.", sold.Name)
		fmt.Printf("[%s] sells a %s to %s for %d gold.\n", npc.Name, sold.Name, merchant.Name, price)

	case models.ActionUse:
		idx := findInventoryIndex(npc, action.Item)
		if idx == -1 || npc.Inventory[idx].Kind != "potion" {
			memoryEntry = fmt.Sprintf("I reached for a %s, but didn't have one ready.", action.Item)
			npc.LastAction = "reached for an item they didn't have."
			break
		}
		potion := npc.Inventory[idx]
		npc.Inventory = append(npc.Inventory[:idx], npc.Inventory[idx+1:]...)
		npc.HP += potion.HealAmount
		if npc.HP > npc.MaxHP {
			npc.HP = npc.MaxHP
		}
		npc.TicksSinceLastHit = 0
		memoryEntry = fmt.Sprintf("I drank a %s and recovered %d HP.", potion.Name, potion.HealAmount)
		npc.LastAction = fmt.Sprintf("used a %s.", potion.Name)
		fmt.Printf("[%s] uses a %s, now at %d/%d HP.\n", npc.Name, potion.Name, npc.HP, npc.MaxHP)

	case models.ActionTalk:
		addressee := action.Target

		if addressee == "" {
			fmt.Printf("[%s] says: \"%s\"\n", npc.Name, action.Dialogue)
			memoryEntry = fmt.Sprintf("I said: \"%s\"", action.Dialogue)
			npc.LastAction = fmt.Sprintf("said: \"%s\"", action.Dialogue)
			w.recordDialogue(npc.Name, "", action.Dialogue)
			break
		}

		other, ok := w.NPCs[addressee]
		canHear := ok && other.HP > 0 && (other.IsLLM || other.IsPlayerControlled) && (npc.IsLLM || npc.IsPlayerControlled)
		if !canHear {
			memoryEntry = fmt.Sprintf("I tried to talk to %s, but they weren't there to listen.", addressee)
			npc.LastAction = fmt.Sprintf("talked to no one — %s wasn't around.", addressee)
			break
		}

		fmt.Printf("[%s] says to %s: \"%s\"\n", npc.Name, other.Name, action.Dialogue)
		w.recordDialogue(npc.Name, other.Name, action.Dialogue)
		memoryEntry = fmt.Sprintf("I said to %s: \"%s\"", other.Name, action.Dialogue)
		npc.LastAction = fmt.Sprintf("talked to %s.", other.Name)

		if npc.IsLLM && other.IsLLM {
			shift := action.TrustShift
			if shift == 0 {
				shift = 3
			} else if shift > 15 {
				shift = 15
			} else if shift < -15 {
				shift = -15
			}
			adjustTrust(npc, other.Name, shift)
			adjustTrust(other, npc.Name, shift)
		}

		if other.IsLLM {
			heard := fmt.Sprintf("%s said to me: \"%s\"", npc.Name, action.Dialogue)
			other.Memory = append(other.Memory, heard)
			if len(other.Memory) > 5 {
				other.Memory = other.Memory[1:]
			}
			if w.Repo != nil {
				w.persistMemory(w.persistCtx(ctx), other, heard)
			}

			other.PendingReplyFrom = npc.Name
			other.PendingReplyText = action.Dialogue
			other.PendingReplySetTurn = w.currentTurn
		}

		if npc.PendingReplyFrom == addressee {
			npc.PendingReplyFrom = ""
			npc.PendingReplyText = ""
		}

	case models.ActionTrade:

		partner, ok := w.NPCs[action.Target]
		if !ok || !partner.IsLLM || partner.HP <= 0 || !w.isAdjacent(npc, partner) {
			memoryEntry = fmt.Sprintf("I tried to complete a deal with %s, but they weren't close enough.", action.Target)
			npc.LastAction = "tried to trade with someone not nearby."
			break
		}
		if action.Gold <= 0 && action.Item == "" {
			memoryEntry = "I tried to finalize a trade, but didn't actually specify gold or an item to hand over."
			npc.LastAction = "fumbled a trade with nothing to give."
			break
		}

		var itemMoved *models.Item
		var itemIdx = -1
		if action.Item != "" {
			itemIdx = findInventoryIndex(npc, action.Item)
			if itemIdx == -1 {
				memoryEntry = fmt.Sprintf("I tried to sell %s a %s, but I don't actually have one.", partner.Name, action.Item)
				npc.LastAction = "tried to trade an item they didn't have."
				break
			}
			itemMoved = npc.Inventory[itemIdx]
		}

		payer, payee := npc, partner
		if itemMoved != nil {
			payer, payee = partner, npc
		}
		if action.Gold > 0 && payer.Gold < action.Gold {
			if itemMoved != nil {
				memoryEntry = fmt.Sprintf("I offered %s to %s for %d gold, but they only have %d — no deal yet.", itemMoved.Name, partner.Name, action.Gold, payer.Gold)
				npc.LastAction = fmt.Sprintf("tried to sell something %s couldn't afford.", partner.Name)
			} else {
				memoryEntry = fmt.Sprintf("I tried to pay %s %d gold, but I only had %d.", partner.Name, action.Gold, npc.Gold)
				npc.LastAction = "couldn't afford to complete a trade."
			}
			break
		}

		if itemMoved != nil {
			npc.Inventory = append(npc.Inventory[:itemIdx], npc.Inventory[itemIdx+1:]...)
			partner.Inventory = append(partner.Inventory, itemMoved)
		}
		if action.Gold > 0 {
			payer.Gold -= action.Gold
			payee.Gold += action.Gold
		}

		var summary strings.Builder
		fmt.Fprintf(&summary, "[%s] completes a trade with %s:", npc.Name, partner.Name)
		if itemMoved != nil {
			fmt.Fprintf(&summary, " a %s", itemMoved.Name)
			if action.Gold > 0 {
				fmt.Fprintf(&summary, " for %d gold (paid by %s)", action.Gold, partner.Name)
			}
		} else if action.Gold > 0 {
			fmt.Fprintf(&summary, " %d gold", action.Gold)
		}
		summary.WriteString(".")
		fmt.Println(summary.String())

		w.recordDialogue(npc.Name, partner.Name, "🤝 "+summary.String())

		memoryEntry = fmt.Sprintf("I completed a trade with %s — settled, nothing more to negotiate about it.", partner.Name)
		npc.LastAction = fmt.Sprintf("traded with %s.", partner.Name)
		partnerMemory := fmt.Sprintf("%s and I completed our trade — settled, nothing more to negotiate about it.", npc.Name)
		partner.Memory = append(partner.Memory, partnerMemory)
		if len(partner.Memory) > 5 {
			partner.Memory = partner.Memory[1:]
		}
		if w.Repo != nil {
			pctx := w.persistCtx(ctx)
			if err := w.Repo.SaveNPC(pctx, partner); err != nil {
				log.Printf("could not persist %s: %v", partner.Name, err)
			}
			w.persistMemory(pctx, partner, partnerMemory)
		}

		adjustTrust(npc, partner.Name, 10)
		adjustTrust(partner, npc.Name, 10)

	default:
		npc.LastAction = "waited."
	}

	if npc.Position != startPos {
		prev := startPos
		npc.PrevPosition = &prev
	}

	if npc.IsLLM {

		if memoryEntry != "" {
			npc.Memory = append(npc.Memory, memoryEntry)
			if len(npc.Memory) > 5 {
				npc.Memory = npc.Memory[1:]
			}
		}

		if w.Repo != nil {
			pctx := w.persistCtx(ctx)
			if err := w.Repo.SaveNPC(pctx, npc); err != nil {
				log.Printf("could not persist %s: %v", npc.Name, err)
			}

			if memoryEntry != "" {
				w.persistMemory(pctx, npc, memoryEntry)
			}
		}
	}
}
