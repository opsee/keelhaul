package store

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/keelhaul/com"
	"strings"
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

func (pg *Postgres) PutRegion(region *com.Region) error {
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

	if err != nil {
		return nil, err
	}

	return &ListBastionsResponse{Bastions: bastions}, nil
}

func (pg *Postgres) putBastion(q sqlx.Queryer, bastion *com.Bastion) error {
	var id string
	err := sqlx.Get(
		q,
		&id,
		`insert into bastions (customer_id, user_id, stack_id, image_id, instance_type, vpc_id, subnet_id, subnet_routing, state, password_hash)
		 values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 returning id`,
		bastion.CustomerID, bastion.UserID, bastion.StackID, bastion.ImageID, bastion.InstanceType,
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

func (pg *Postgres) putRegion(x sqlx.Ext, region *com.Region) error {
	data, err := json.Marshal(region)
	if err != nil {
		return err
	}

	insert := map[string]interface{}{
		"customer_id": region.CustomerID,
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

func in(ordStart, listLen int) string {
	ords := make([]string, listLen)
	for i := 0; i < listLen; i++ {
		ords[i] = fmt.Sprintf("$%d", i+ordStart)
	}

	return strings.Join(ords, ",")
}
