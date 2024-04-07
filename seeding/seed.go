package seeding

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// RunSeed : takes in the reader and seeder, will eventually push the json data into the database
// Gets the conditional seeding flag from the environment.  - no,ifempty,force
// force will flush the database and then push the data in
// ifempty will push the data only if database is empty
// no will not push or read any data
func RunSeed(rdr SeedReader, sdr Seeder) {
	flag := os.Getenv("DB_SEED")
	if flag == "force" || flag == "ifempty" {
		if flag == "force" {
			if err := sdr.Flush(); err != nil {
				log.Errorf("Error flushing the database, %s", err)
				return
			}
		} else if !sdr.DestinationEmpty() {
			log.Error("Destination database isnt empty for ifempty setting")
			return
		}
		// If the code gets to this, the database is empty and ready to be populated.
		count, err := sdr.Seed(rdr)
		if err != nil {
			log.Errorf("failed to seed database %s", err)
			return
		}
		sdr.Close()
		log.WithFields(log.Fields{
			"count": count,
		}).Info("seeding done")
		return
	}
}
