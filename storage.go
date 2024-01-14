package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

// GetAccounts implements Storage.

func NewPostgresStore() (*PostgresStore, error) {
	godotenv.Load()
	var (
		dbname     = os.Getenv("DB_NAME")
		dbuser     = os.Getenv("DB_USER")
		dbpassword = os.Getenv("DB_PASSWORD")
		dbhost     = os.Getenv("DB_HOST")
		dbport     = os.Getenv("DB_PORT")
		uri        = fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable", dbuser, dbname, dbpassword, dbhost, dbport)
	)

	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, fmt.Errorf("opening sql\n %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sql\n %v", err)
	}

	return &PostgresStore{
		db: db,
	}, nil

}

func (s *PostgresStore) init() error {
	return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {

	createTableSQL := `
    	create table if not exists account (
 	  		id serial primary key,
  	  		first_name varchar(50) not null,
   			last_name varchar(50) not null,
  			account_number int,
 			balance decimal(10, 2),
  			created_at timestamp
		);`
	/*
		Use Exec when you are executing SQL statements that don't return rows,
			such as INSERT, UPDATE, DELETE, etc
	*/
	_, err := s.db.Exec(createTableSQL)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {

	query := `
		insert into account(
			first_name,
			last_name, 
			account_number, 
			balance, 
			created_at
		)
        values ($1, $2, $3, $4, $5)		
	`

	_, err := s.db.Exec(
		query,
		acc.FirstName,
		acc.LastName,
		acc.AccoutnNumber,
		acc.Balance,
		acc.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(a *Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	_, err := s.db.Exec("delete from account where id = $1", id)
	return err
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {

	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, err
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {

	rows, err := s.db.Query("select * from account where id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account id %d not found", id)
}

// Get Account Details
func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	if err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.AccoutnNumber,
		&account.Balance,
		&account.CreatedAt,
	); err != nil {
		return nil, err
	}

	return account, nil
}
