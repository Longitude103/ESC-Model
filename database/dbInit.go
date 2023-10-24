package database

import (
	"github.com/jmoiron/sqlx"
)

// InitializeDb creates the database results table if it doesn't exist, so the records of the
// transaction can be stored properly, also creates file_keys table for result file_type integer
// and a foreign key restriction to results table
func InitializeDb(db *sqlx.DB, logger lg) error {
	stmt, err := db.Prepare(`
		create table if not exists file_keys
			(
				file_key integer not null
					constraint file_keys_pk
						primary key autoincrement,
				description TEXT
			);
			
		create unique index file_keys_file_key_uindex
				on file_keys (file_key);
	`)
	if err != nil {
		logger.Errorf("Error in create statement for file keys: %s", err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		logger.Errorf("Error on create file_keys: %s", err)
		return err
	}

	// if table exists with records, then skip adding data.
	var count int
	err = db.QueryRow(`select count(*) from file_keys;`).Scan(&count)
	if err != nil {
		logger.Errorf("Error in count file_key records %s", err)
		return err
	}

	if count == 0 {
		stmt, err = db.Prepare(`
		insert into file_keys (file_key, description) values (101, 'Dryland'), (102, 'Natural Vegetation'), 
		  (103, 'NP Surface Water Only'), (104, 'SP Surface Water Only'), (105, 'NP Comingled Parcel Pre 1998'), 
		  (106, 'NP Comingled Parcel Post 1997'), (107, 'SP Comingled Parcel Pre 1998'), (108, 'SP Comingled Parcel Post 1997'), 
		  (109, 'NP Groundwater Only Pre 1998'), (110, 'NP Groundwater Only Post 1997'), (111, 'SP Groundwater Only Pre 1998'), 
		  (112, 'SP Groundwater Only Post 1997'), (113, 'NP Canal Loss'), (114, 'SP Canal Loss'), (115, 'Outside NP and SP'), 
		  (116, 'NP Recharge Sites'), (117, 'SP Recharge Sites'), (118, 'Colorado'), (119, 'Wyoming'), 
		  (120, 'UNWNRD'), (121, 'CoHyst-FA'), (122, 'CoHyst-25'), (123, 'CoHyst-Western'), (124, 'NP Excess Flow'), (201, 'NP Comingled Pre 1998'), 
		  (202, 'NP Groundwater Only Pre 1998'), (203, 'NP Comingled Post 1997'), (204, 'NP Groundwater Only Post 1997'), 
		  (205, 'SP Comingled Pre 1998'), (206, 'SP Groundwater Only Pre 1998'), (207, 'SP Comingled Post 1997'), (208, 'SP Groundwater Only Post 1997'), 
          (209, 'Steady State'), (210, 'NP Municipal'), (211, 'NP Industrial'), (212, 'Other Wells'), (213, 'Western Canal Outside SP'),
          (214, 'Colorado'), (215, 'Wyoming'), (216, 'UNWNRD'), (217, 'CoHyst-FA'), (218, 'CoHyst-25'), (219, 'SP Municipal'), (220, 'SP Industrial'), (221, 'Un-retired Well Pumping')`)
		if err != nil {
			logger.Errorf("Error in statement of key records: %s", err)
			return err
		}

		_, err = stmt.Exec()
		if err != nil {
			logger.Errorf("Error in insert of key records: %s", err)
			return err
		}
	}

	stmt, err = db.Prepare(`
		create table if not exists results
			(
				id integer not null
					constraint results_pk
						primary key autoincrement,
				cell_node int not null,
				cell_size float,
				dt TIMESTAMP,
				file_type int not null
					constraint results_file_keys_file_key_fk
					references file_keys,
				result float
			);
		
		create unique index results_id_uindex
			on results (id);
	`)
	if err != nil {
		logger.Errorf("Error in creating results table statement: %s", err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		logger.Errorf("Error in creating results table: %s", err)
		return err
	}

	// add results table for parcelNIR
	stmt, err = db.Prepare(`create table if not exists parcelNIR
									(
										parcelID integer,
										nrd text,
										dt TIMESTAMP,
										nir real,
										irrtype integer
									);`)
	if err != nil {
		logger.Errorf("Error in statement of parcel nir table: %s", err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		logger.Errorf("Error in creating parcel nir table: %s", err)
		return err
	}

	stmt, err = db.Prepare(`create table if not exists swDelivery
									(
										canalId integer,
										dt TIMESTAMP,
										delAmount real
									);`)
	if err != nil {
		logger.Errorf("Error in statement of SW Delivery table: %s", err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		logger.Errorf("Error in creating SW Delivery table: %s", err)
		return err
	}

	// add results table for parcelPumping
	stmt, err = db.Prepare(`create table if not exists parcelPumping
									(
										parcelID integer,
										nrd text,
										dt TIMESTAMP,
										file_type int not null
											constraint parcelPumping_file_keys_file_key_fk
											references file_keys,
										pump real
									);`)
	if err != nil {
		logger.Errorf("Error in statement of parcel pumping table: %s", err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		logger.Errorf("Error in creating parcel pumping table: %s", err)
		return err
	}

	stmt, err = db.Prepare(`
		create table if not exists wel_results
			(
				id integer not null
					constraint results_pk
						primary key autoincrement,
				well_id int not null,
				cell_node int not null,
				dt TIMESTAMP,
				file_type int not null
					constraint results_file_keys_file_key_fk
					references file_keys,
				result float
			);
		
		create unique index results_id_uindex
			on wel_results (id);
	`)
	if err != nil {
		logger.Errorf("Error in creating wel_results table statement: %s", err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		logger.Errorf("Error in creating wel_results table: %s", err)
		return err
	}

	stmt, err = db.Prepare(`
		create table if not exists results_notes 
			(
			    id integer not null
					constraint results_notes_pk
						primary key autoincrement,
				note text
			)
    `)
	if err != nil {
		logger.Errorf("Error in creating results_notes table statement: %s", err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		logger.Errorf("Error in creating results_notes table: %s", err)
		return err
	}

	stmt, err = db.Prepare(`
		create table if not exists cellrc 
			(
			    node integer not null
					constraint cellrc_pk
						primary key,
				rw integer,
				clm integer
			)
    `)

	if err != nil {
		logger.Errorf("Error in creating cellrc table statement: %s", err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		logger.Errorf("Error in creating results_notes table: %s", err)
		return err
	}

	return nil
}
