package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type parsedSkill struct {
	Name       string
	TargetName string
	Ability    string
	AbilityID  int64
	Flags      int64
	Damage     int64
	Crit       bool
	Direct     bool
}

type parseReader struct {
	parseFile   *os.File
	playerLines []string
	skills      []parsedSkill
}

func (p *parseReader) loadParseFile(fileName string) {
	p.parseFile = loadTextFile(fileName)
	scanner := bufio.NewScanner(p.parseFile)
	p.playerLines = make([]string, 0)
	p.skills = make([]parsedSkill, 0)
	for scanner.Scan() {
		thisLine := scanner.Text()
		if strings.Contains(thisLine, players[0].Name) {
			sp := strings.Split(thisLine, "|")
			if (sp[0] == "21" || sp[0] == "22") && sp[3] != sp[7] {
				fmt.Println(thisLine)
				p.playerLines = append(p.playerLines, thisLine)
				AbilityID, _ := strconv.ParseInt(sp[4], 16, 0)
				Flags, _ := strconv.ParseInt(sp[8], 16, 0)
				Damage, _ := strconv.ParseInt(sp[9], 16, 0)
				Damage = Damage >> 16
				Crit := (Flags & 0x100) != 0
				Direct := (Flags & 0x200) != 0
				skill := parsedSkill{sp[3], sp[7], sp[5], AbilityID, Flags, Damage, Crit, Direct}
				p.skills = append(p.skills, skill)
				skill.log()
			}
			// Need to handle below message as well
			/* sp[0] == "24" || sp[0] == "26" || sp[0] == "30"*/
		}
	}
}

func (p *parsedSkill) log() {
	globalLog(Important, "Name: "+p.Name+"->"+p.TargetName)
	globalLog(Important, "Skill Used: "+strconv.FormatInt(p.AbilityID, 16)+" - "+p.Ability)
	globalLog(Important, "Crit: "+strconv.FormatBool(p.Crit)+" Direct: "+strconv.FormatBool(p.Direct))
	globalLog(Important, "Damage: "+strconv.FormatInt(p.Damage, 10))
	globalLog(Important, "")
}
