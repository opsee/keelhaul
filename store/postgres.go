package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	log "github.com/opsee/logrus"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(connection string) (Store, error) {
	db, err := sqlx.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(8)

	return &Postgres{
		db: db,
	}, nil
}

func (pg *Postgres) PutBastion(bastion *com.Bastion) error {
	return pg.putBastion(pg.db, bastion)
}

func (pg *Postgres) UpdateBastion(bastion *com.Bastion) error {
	return pg.updateBastion(pg.db, bastion)
}

func (pg *Postgres) PutRegion(region *schema.Region) error {
	return pg.putRegion(pg.db, region)
}

func (pg *Postgres) GetBastion(request *GetBastionRequest) (*GetBastionResponse, error) {
	bastion := &com.Bastion{}
	err := pg.db.Get(
		bastion,
		"select * from bastions where id = $1 and state = $2",
		request.ID,
		request.State,
	)

	if err != nil {
		return nil, err
	}

	return &GetBastionResponse{Bastion: bastion}, nil
}

func (pg *Postgres) ListBastions(request *ListBastionsRequest) (*ListBastionsResponse, error) {
	query := fmt.Sprintf("select * from bastions where customer_id = $1 and state in (%s)", in(2, len(request.State)))
	args := make([]interface{}, len(request.State)+1)
	args[0] = request.CustomerID
	for i, s := range request.State {
		args[i+1] = s
	}

	bastions := make([]*com.Bastion, 0)
	err := pg.db.Select(
		&bastions,
		query,
		args...,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &ListBastionsResponse{Bastions: bastions}, nil
}

func (pg *Postgres) putBastion(q sqlx.Queryer, bastion *com.Bastion) error {
	var id string
	err := sqlx.Get(
		q,
		&id,
		`insert into bastions (customer_id, user_id, stack_id, image_id, instance_type, region, vpc_id, subnet_id, subnet_routing, state, password_hash)
		 values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 returning id`,
		bastion.CustomerID, bastion.UserID, bastion.StackID, bastion.ImageID, bastion.InstanceType, bastion.Region,
		bastion.VPCID, bastion.SubnetID, bastion.SubnetRouting, bastion.State, bastion.PasswordHash,
	)

	bastion.ID = id
	return err
}

func (pg *Postgres) updateBastion(x sqlx.Ext, bastion *com.Bastion) error {
	_, err := sqlx.NamedExec(
		x,
		`update bastions set stack_id = :stack_id, image_id = :image_id,
		 instance_id = :instance_id, group_id = :group_id, state = :state where id = :id`,
		bastion,
	)

	return err
}

func (pg *Postgres) putRegion(x sqlx.Ext, region *schema.Region) error {
	data, err := json.Marshal(region)
	if err != nil {
		return err
	}

	insert := map[string]interface{}{
		"customer_id": region.CustomerId,
		"region":      region.Region,
		"data":        data,
	}

	_, err = sqlx.NamedExec(
		x,
		`with update_regions as (update regions set (data) = (:data) where region = :region
		 and customer_id = :customer_id returning region),
		 insert_regions as (insert into regions (region, customer_id, data)
		 select :region as region, :customer_id as customer_id, :data as data
		 where not exists (select region from update_regions limit 1) returning region)
		 select * from update_regions union all select * from insert_regions`,
		insert,
	)

	return err
}

func (pg *Postgres) UpdateTrackingSeen(bastionIDs []string, customerIDs []string) error {
	for i, s := range bastionIDs {
		bastionIDs[i] = fmt.Sprintf("cast('%s' as UUID)", s)
		customerIDs[i] = fmt.Sprintf("cast('%s' as UUID)", customerIDs[i])
	}
	query := fmt.Sprintf("select batch_upsert_tracking(array[%s], array[%s])", strings.Join(bastionIDs, ", "), strings.Join(customerIDs, ", "))
	// TODO consider using prepared stmt w/placeholders
	r, err := pg.db.Query(query)
	if err == nil {
		r.Close()
	}
	return err
}

func in(ordStart, listLen int) string {
	ords := make([]string, listLen)
	for i := 0; i < listLen; i++ {
		ords[i] = fmt.Sprintf("$%d", i+ordStart)
	}

	return strings.Join(ords, ",")
}

func (pg *Postgres) ListTrackingStates(offset int, limit int) (*TrackingStateResponse, error) {
	query := "select id,customer_id,status,last_seen from bastion_tracking order by id"
	if limit > 0 {
		query += fmt.Sprintf(" limit %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" offset %d", offset)
	}

	states := make([]*TrackingState, 0)
	args := make([]interface{}, 0)
	err := pg.db.Select(&states, query, args...)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &TrackingStateResponse{States: states}, nil
}

func (pg *Postgres) ListBastionStates(customers []string, filters ...*opsee.Filter) (*TrackingStateResponse, error) {
	if len(customers) == 0 {
		query := "select bastion_tracking.id,bastion_tracking.customer_id,bastion_tracking.status,bastion_tracking.last_seen,bastions.region,bastions.vpc_id from bastion_tracking inner join bastions on (bastion_tracking.id = bastions.id)"
		if len(filters) > 0 {
			query += " WHERE 1=1 " // no-op to avoid AND logic
			for _, filter := range filters {
				switch filter.Key {
				case "status":
					query += fmt.Sprintf("AND bastion_tracking.status = '%s'", filter.Value)
				case "region":
					query += fmt.Sprintf("AND bastions.region = '%s'", filter.Value)
				case "customer_id":
					query += fmt.Sprintf("AND bastions.customer_id = '%s'", filter.Value)
				case "vpc_id":
					query += fmt.Sprintf("AND bastions.vpc_id = '%s'", filter.Value)
				}
			}
			query += ";"
			log.Debugf("Created filtered query: %s", query)

		}
		states := make([]*TrackingState, 0)
		args := make([]interface{}, 0)

		err := pg.db.Select(&states, query, args...)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		return &TrackingStateResponse{States: states}, nil
	}
	query := "select id,customer_id,status,last_seen from bastion_tracking where customer_id in"
	casted := make([]string, 0, len(customers))
	for _, b := range customers {
		casted = append(casted, fmt.Sprintf("cast('%s' as uuid)", b))
	}
	custSet := strings.Join(casted, ", ")

	states := make([]*TrackingState, 0)
	args := make([]interface{}, 0)
	q := fmt.Sprintf("%s (%s)", query, custSet)
	err := pg.db.Select(&states, q, args...)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &TrackingStateResponse{States: states}, nil
}

func (pg *Postgres) GetPendingTrackingStates(inactiveInterval string) (*TrackingStateResponse, error) {
	query := fmt.Sprintf("select id,customer_id,status,last_seen from bastion_tracking "+
		"where (status = 'active' and last_seen <= (now() - interval '%s')) "+
		"or (status = 'inactive' and last_seen >= (now() - interval '%s'))",
		inactiveInterval, inactiveInterval)

	states := make([]*TrackingState, 0)
	args := make([]interface{}, 0)
	err := pg.db.Select(&states, query, args...)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &TrackingStateResponse{States: states}, nil
}

func (pg *Postgres) UpdateTrackingState(bastionID string, newState string) error {
	stmt := fmt.Sprintf("update bastion_tracking set status = '%s' "+
		"where id = cast('%s' as UUID)",
		newState, bastionID)

	res, err := pg.db.Exec(stmt)
	if err == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			if err == nil {
				err = errors.New("no rows updated")
			}
		}
	}

	return err
}
