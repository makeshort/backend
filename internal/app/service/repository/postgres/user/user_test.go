package user

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"log"
	"testing"
)

func TestUser_Create(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sqlx.DB) {
		err = db.Close()
		if err != nil {

		}
	}(db)

	r := New(db)

	type args struct {
		email        string
		username     string
		passwordHash string
	}
	type mockBehavior func(args args, id string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		want         string
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				email:        "some email",
				username:     "some password",
				passwordHash: "some password hash",
			},
			want: "some uuid",
			mockBehavior: func(args args, id string) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)

				mock.ExpectQuery("INSERT INTO users").
					WithArgs(args.email, args.username, args.passwordHash).WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Empty fields",
			args: args{
				email:        "",
				username:     "",
				passwordHash: "",
			},
			mockBehavior: func(args args, id string) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id).RowError(1, errors.New("some error"))

				mock.ExpectQuery("INSERT INTO users").
					WithArgs(args.email, args.username, args.passwordHash).WillReturnRows(rows).WillReturnError(errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.args, tc.want)

			got, err := r.Create(context.Background(), tc.args.email, tc.args.username, tc.args.passwordHash)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, got, tc.want)
			}
		})
	}
}

func TestUser_Delete(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sqlx.DB) {
		err = db.Close()
		if err != nil {

		}
	}(db)

	r := New(db)

	type args struct {
		uuid string
	}
	type mockBehavior func(args args, id string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		want         string
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				uuid: "some uuid",
			},
			want: "some uuid",
			mockBehavior: func(args args, id string) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)

				mock.ExpectQuery("DELETE FROM users").
					WithArgs(args.uuid).WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Empty fields",
			args: args{
				uuid: "",
			},
			mockBehavior: func(args args, id string) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id).RowError(1, errors.New("some error"))

				mock.ExpectQuery("DELETE FROM users").
					WithArgs(args.uuid).WillReturnRows(rows).WillReturnError(errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.args, tc.want)

			err := r.Delete(context.Background(), tc.args.uuid)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
