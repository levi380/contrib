package helper

import (
	"fmt"
	"github.com/yitter/idgenerator-go/idgen"
)



func Initial(workid uint16) {
	var options = idgen.NewIdGeneratorOptions(workid)
	options.WorkerIdBitLength = 12
	idgen.SetIdGenerator(options)
}

func GenId() string {

	id := idgen.NextId()

	return fmt.Sprintf("%d", id)
}
