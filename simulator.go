package main

import (
	"strconv"
)

const timeIncrement float64 = 0.01

type simulator struct {
	Players []*playerCharacter
	Enemies []*enemy
}

type simResult struct {
	skillCount      int
	critCount       int
	directCount     int
	critDirectCount int
	totalDamageDone int
	timeEllapsed    float64
	skillsPerformed []int
	autoAttacks     int
	dotTicks        int
}

func (s *simulator) RunSkillQueue(SkillQueue []string) simResult {
	currentSkillIndex := 0
	var totalResult simResult
	totalResult.skillsPerformed = make([]int, len(s.Players[0].Skills))

	for currentSkillIndex < len(SkillQueue) {
		for i := range s.Players {
			results := enemies[0].advanceTime(timeIncrement)
			if results != nil {
				totalResult.dotTicks += len(results)
			}

			// First advance time to take care of Dots, Cooldowns, and Autos
			timeResult := s.Players[i].advanceTime(timeIncrement, enemies[0])

			if timeResult != nil {
				totalResult.autoAttacks++
				if results != nil {
					results = append(results, *timeResult)
				} else {
					results = []skillResult{*timeResult}
				}
			}
			for _, result := range results {
				enemies[0].applyDamage(&result)
				result.Log()
				result.ApplyBuffs(*s.Players[i])
				totalResult.totalDamageDone += result.DamageDone
				if result.DidCrit {
					totalResult.critCount++
				}
				if result.DidDirect {
					totalResult.directCount++
				}
				if result.DidCrit && result.DidDirect {
					totalResult.critDirectCount++
				}
			}
			totalResult.timeEllapsed += timeIncrement

			// Next attempt to perform the next action in the queue
			actionResult := s.Players[i].performAction(SkillQueue[currentSkillIndex], enemies[0])
			if actionResult != nil {
				enemies[0].applyDamage(actionResult)
				actionResult.Log()
				actionResult.ApplyBuffs(*s.Players[i])
				totalResult.totalDamageDone += actionResult.DamageDone
				totalResult.skillsPerformed[actionResult.SkillIndex]++
				currentSkillIndex++
				totalResult.skillCount++
				if actionResult.DidCrit {
					totalResult.critCount++
				}
				if actionResult.DidDirect {
					totalResult.directCount++
				}
				if actionResult.DidCrit && actionResult.DidDirect {
					totalResult.critDirectCount++
				}
			}
		}
	}

	return totalResult
}

func (s *simulator) RunParse(parsedSkills []parsedSkill) simResult {
	var totalResult simResult
	totalResult.skillsPerformed = make([]int, len(s.Players[0].Skills))
	currentSkillIndex := 0
	simErrors := 0
	validSkills := 0
	invalidSkills := 0

	for _, skill := range parsedSkills {
		if skill.Ability == "Attack" {
			continue
		}
		skillPerformed := false
		for skillPerformed == false {
			if totalResult.timeEllapsed >= skill.RelativeTime && skill.Ability != "Attack" {
				// Perform current action
				actionResult := s.Players[0].performAction(skill.Ability, enemies[0])
				if actionResult != nil {
					enemies[0].applyDamage(actionResult)
					actionResult.Log()
					actionResult.ApplyBuffs(*s.Players[0])
					totalResult.totalDamageDone += actionResult.DamageDone
					totalResult.skillsPerformed[actionResult.SkillIndex]++
					currentSkillIndex++
					totalResult.skillCount++
					if actionResult.DidCrit {
						totalResult.critCount++
					}
					if actionResult.DidDirect {
						totalResult.directCount++
					}
					if actionResult.DidCrit && actionResult.DidDirect {
						totalResult.critDirectCount++
					}
				} else {
					globalLog(Important, "SKILL UNAVAILABLE ERROR: "+skill.Ability)
					simErrors++
				}

				skillPerformed = true
				// Reset timing, as the logs may have latency in them.
				s.Players[0].GCDTimeRemaining = 0
				s.Players[0].IsGCDOnCooldown = false
				s.Players[0].AnimationLock = 0

				// Check Damage
				if skill.Crit == actionResult.DidCrit && skill.Direct == actionResult.DidDirect {
					damage := actionResult.BaseDamageDone
					if actionResult.DidCrit {
						damage *= actionResult.CritMultiplier
					}
					if actionResult.DidDirect {
						damage *= 1.2
					}
					lowerBound := float64(damage) * 0.95
					upperBound := float64(damage) * 1.05
					lowerBound *= actionResult.DamageBuff
					upperBound *= actionResult.DamageBuff
					if float64(skill.Damage) >= lowerBound && float64(skill.Damage) <= upperBound {
						globalLog(Basic, "SKILL VALIDATED")
						skill.Validated = true
						validSkills++
					} else {
						globalLog(Important, "INVALID DAMAGE")
						globalLog(Important, "Simulated: "+strconv.Itoa(actionResult.DamageDone))
						globalLog(Important, "Parse: "+strconv.FormatInt(skill.Damage, 10))
						invalidSkills++
					}

				}
			}

			// Advance time to take care of Dots, Cooldowns, and Autos
			results := enemies[0].advanceTime(timeIncrement)
			if results != nil {
				totalResult.dotTicks += len(results)
				for i := range results {
					totalResult.totalDamageDone += results[i].DamageDone
				}
			}

			timeResult := s.Players[0].advanceTime(timeIncrement, enemies[0])

			if timeResult != nil {
				totalResult.totalDamageDone += timeResult.DamageDone
				if timeResult.SkillName == "Attack" {
					totalResult.autoAttacks++
				}
				// TODO Figure out proper logic here
				//if skill.Ability == "Attack" {
				//	if s.Players[0].TotalTime > skill.RelativeTime+0.1 {
				//		globalLog(Important, "SKILL TIMING ERROR: "+skill.Ability)
				//	}
				//	skillPerformed = true
				//} else {
				//	globalLog(Important, "AUTO ATTACK TIMING")
				//}
				if results != nil {
					results = append(results, *timeResult)
				} else {
					results = []skillResult{*timeResult}
				}
			}
			for _, result := range results {
				enemies[0].applyDamage(&result)
				result.Log()
				result.ApplyBuffs(*s.Players[0])
				if result.DidCrit {
					totalResult.critCount++
				}
				if result.DidDirect {
					totalResult.directCount++
				}
				if result.DidCrit && result.DidDirect {
					totalResult.critDirectCount++
				}
			}
			totalResult.timeEllapsed += timeIncrement
		}
	}

	globalLog(Important, "Total Errors: "+strconv.Itoa(simErrors))
	globalLog(Important, "Validated Skills: "+strconv.Itoa(validSkills))
	globalLog(Important, "Invalid Skills: "+strconv.Itoa(invalidSkills))

	return totalResult
}

func (s *simResult) log() {
	globalLog(Important, "Total Damage Done: "+strconv.Itoa(s.totalDamageDone))
	globalLogFloat(Important, "Time ellapsed: ", s.timeEllapsed)
	globalLogFloat(Important, "DPS: ", float64(s.totalDamageDone)/s.timeEllapsed)
	globalLog(Important, "Skills performed: "+strconv.Itoa(s.skillCount))
	globalLogFloat(Important, "Crit Rate: ", float64(s.critCount)/float64(s.skillCount))
	globalLogFloat(Important, "Direct Rate: ", float64(s.directCount)/float64(s.skillCount))
	globalLogFloat(Important, "CritDirect Rate: ", float64(s.critDirectCount)/float64(s.skillCount))
	globalLogIntSlice(Important, "Skill Counts: ", s.skillsPerformed)
	globalLogFloat(Important, "Auto Attacks: ", float64(s.autoAttacks))
	globalLogFloat(Important, "DoT Ticks: ", float64(s.dotTicks))
	globalLog(Important, "")
}
