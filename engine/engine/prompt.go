package engine

import (
	"ai-agent-engine/models"
	"fmt"
)

func (w *World) BuildPerception(npc *models.NPC, s Surroundings, memories []string) NPCPerception {
	p := NPCPerception{
		Name:             npc.Name,
		Role:             npc.Role,
		Personality:      orDefault(npc.Personality, "Not strongly defined — improvise something consistent with your Background."),
		Faction:          orDefault(npc.Faction, "None in particular."),
		CurrentMood:      orDefault(npc.CurrentMood, "Neutral."),
		Background:       npc.Background,
		Objective:        npc.Objective,
		Relationships:    npc.Relationships,
		Likes:            npc.Likes,
		Dislikes:         npc.Dislikes,
		KnownSecrets:     npc.KnownSecrets,
		Level:            npc.Level,
		HP:               npc.HP,
		MaxHP:            npc.MaxHP,
		BaseAttack:       npc.BaseAttack,
		CombatStyle:      orDefault(npc.CombatStyle, "Unspecified."),
		Skills:           npc.Skills,
		Gold:             npc.Gold,
		Position:         npc.Position,
		NearbyPOIs:       s.NearbyPOIs,
		RelevantMemories: memories,
	}

	if npc.GoalPosition != nil {
		p.CurrentGoal = fmt.Sprintf("[%d, %d]", npc.GoalPosition.X, npc.GoalPosition.Y)
	}

	if npc.Equipped != nil {
		p.WeaponInfo = &WeaponInfo{
			Name:          npc.Equipped.Name,
			Power:         npc.Equipped.Power,
			Durability:    npc.Equipped.Durability,
			MaxDurability: npc.Equipped.MaxDurability,
		}
	}

	for _, it := range npc.Inventory {
		p.InventoryNames = append(p.InventoryNames, it.Name)
	}

	for _, other := range w.NPCs {
		if !other.IsMerchant || w.distance(npc.Position, other.Position) > npc.VisionRadius {
			continue
		}
		var stock []string
		for _, it := range other.Shop {
			stock = append(stock, fmt.Sprintf("%s (%d gold)", it.Name, it.Price))
		}
		p.NearbyMerchant = &MerchantInfo{Name: other.Name, Stock: stock}
		break
	}

	if s.CurrentTile != nil {
		p.Biome = s.CurrentTile.Biome
		p.TileDescription = s.CurrentTile.Description
		p.IsSafeZone = s.CurrentTile.IsSafeZone
	}

	p.VisibleEntities = s.Descriptions

	if len(s.Descriptions) > 0 {
		for _, other := range w.NPCs {
			if other.Name == npc.Name || w.distance(npc.Position, other.Position) > npc.VisionRadius {
				continue
			}
			trust, hasTrust := npc.Trust[other.Name]
			fear, hasFear := npc.Fear[other.Name]
			if hasTrust || hasFear {
				p.TrustFearLines = append(p.TrustFearLines, fmt.Sprintf("%s: trust %d/100, fear %d/100", other.Name, trust, fear))
			}
		}
	}

	if s.AdjacentMonster != nil {
		foeLabel := s.AdjacentMonster.Species
		if foeLabel == "" {
			foeLabel = s.AdjacentMonster.Name
		}
		p.Alerts.AdjacentMonster = &AdjacentMonsterAlert{
			SpeciesOrName: foeLabel,
			InSafeZone:    s.CurrentTile != nil && s.CurrentTile.IsSafeZone,
		}
	}

	if npc.NegotiationTicks >= 2 && npc.NegotiationPartner != "" {
		for _, v := range s.VisibleNames {
			if v == npc.NegotiationPartner {
				p.Alerts.NegotiationStalled = &NegotiationAlert{
					Partner: npc.NegotiationPartner,
					Turns:   npc.NegotiationTicks,
				}
				break
			}
		}
	}

	if npc.PendingReplyFrom != "" {
		for _, v := range s.VisibleNames {
			if v == npc.PendingReplyFrom {
				p.Alerts.PendingReply = &PendingReplyAlert{
					From: npc.PendingReplyFrom,
					Text: npc.PendingReplyText,
				}
				break
			}
		}
	}

	return p
}
