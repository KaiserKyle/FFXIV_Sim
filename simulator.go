package main

import (
	"strconv"
)

const timeIncrement float64 = 0.01

type simulator struct {
	Players    []*playerCharacter
	Enemies    []*enemy
	SkillQueue []string
}

func (s *simulator) Run() {
	currentSkillIndex := 0
	critCount := 0
	directCount := 0
	critDirectCount := 0
	totalDamageDone := 0
	timeEllapsed := 0.0

	for currentSkillIndex < len(s.SkillQueue) {
		for i := range s.Players {
			enemies[0].advanceTime(timeIncrement)
			result := s.Players[i].advanceTime(timeIncrement, enemies[0])
			if result != nil {
				enemies[0].applyDamage(result)
				result.Log()
				result.ApplyBuffs(*s.Players[i])
				totalDamageDone += result.DamageDone
				if result.DidCrit {
					critCount++
				}
				if result.DidDirect {
					directCount++
				}
				if result.DidCrit && result.DidDirect {
					critDirectCount++
				}
			}
			timeEllapsed += timeIncrement
			result = s.Players[i].performAction(s.SkillQueue[currentSkillIndex], enemies[0])
			if result != nil {
				enemies[0].applyDamage(result)
				result.Log()
				result.ApplyBuffs(*s.Players[i])
				totalDamageDone += result.DamageDone
				currentSkillIndex++
				if result.DidCrit {
					critCount++
				}
				if result.DidDirect {
					directCount++
				}
				if result.DidCrit && result.DidDirect {
					critDirectCount++
				}
			}
		}
	}

	globalLog(Important, "Total Damage Done: "+strconv.Itoa(totalDamageDone))
	globalLogFloat(Important, "Time ellapsed: ", timeEllapsed)
	globalLogFloat(Important, "DPS: ", float64(totalDamageDone)/timeEllapsed)
	globalLog(Important, "Skills performed: "+strconv.Itoa(currentSkillIndex))
	globalLogFloat(Important, "Crit Rate: ", float64(critCount)/float64(currentSkillIndex))
	globalLogFloat(Important, "Direct Rate: ", float64(directCount)/float64(currentSkillIndex))
	globalLogFloat(Important, "CritDirect Rate: ", float64(critDirectCount)/float64(currentSkillIndex))
}
