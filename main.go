package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v5"
	"os"
)

type DupRepository struct {
	Name  string
	Count int64
}

func main() {

	dbURL := flag.String("dburl", "postgres://postgres:root123@localhost:5432/registry", "database url")
	command := flag.String("command", "list", "command to run, can be list or fix")

	flag.Usage = func() {
		fmt.Println("Fix duplicate repository name before reindex repository, this is a one time script\n" +
			"Please make sure you have a backup of the database before running this script!!!")
	}
	flag.Parse()

	if *dbURL == "" {
		flag.Usage()
		os.Exit(1)
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, *dbURL)
	defer conn.Close(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	repos, err := DuplicateRepositories(ctx, conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// dump the repos
	for _, r := range repos {
		fmt.Printf("Found duplicated repository name:%s, count:%d\n", r.Name, r.Count)
	}

	if len(repos) == 0 {
		fmt.Println("No duplicate repository name found!")
	}

	if *command == "list" {
		return
	}

	if *command == "fix" {
		for _, r := range repos {
			newRepoID, oldRepoID, err := FetchDuplicateRepositories(ctx, conn, r.Name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = MigrateTag(ctx, conn, newRepoID, oldRepoID)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = MigrateArtifact(ctx, conn, newRepoID, oldRepoID)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = DeleteDuplicateRepositories(ctx, conn, oldRepoID)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		fmt.Println("Fix done! please run this command in database:\n REINDEX TABLE repository;")

	}
}

func DuplicateRepositories(ctx context.Context, conn *pgx.Conn) ([]*DupRepository, error) {
	var dups []*DupRepository
	rows, err := conn.Query(ctx, `
		SELECT name, count(*) cnt
		FROM repository
		GROUP BY name
		HAVING COUNT(*) > 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var cnt int64
		err := rows.Scan(&name, &cnt)
		if err != nil {
			return nil, err
		}
		dups = append(dups, &DupRepository{
			Name:  name,
			Count: cnt,
		})
	}
	return dups, nil
}

func DeleteDuplicateRepositories(ctx context.Context, conn *pgx.Conn, repoID int64) error {
	_, err := conn.Exec(ctx, `
		DELETE FROM repository
		WHERE repository_id = $1
	`, repoID)
	return err
}

func MigrateTag(ctx context.Context, conn *pgx.Conn, newRepoID, oldRepoID int64) error {
	_, err := conn.Exec(ctx, `
		UPDATE tag
		SET repository_id = $1
		WHERE repository_id = $2
	`, newRepoID, oldRepoID)
	return err
}

func MigrateArtifact(ctx context.Context, conn *pgx.Conn, newRepoID, oldRepoID int64) error {
	_, err := conn.Exec(ctx, `
		UPDATE artifact
		SET repository_id = $1
		WHERE repository_id = $2
	`, newRepoID, oldRepoID)
	return err
}

func FetchDuplicateRepositories(ctx context.Context, conn *pgx.Conn, name string) (int64, int64, error) {
	var repoIDs []int64
	rows, err := conn.Query(ctx, `
		SELECT repository_id
		FROM repository
		WHERE name like $1 ORDER BY creation_time desc`, "%"+name+"%")
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var repoID int64
		err := rows.Scan(&repoID)
		if err != nil {
			return 0, 0, err
		}
		repoIDs = append(repoIDs, repoID)
	}

	if len(repoIDs) != 2 {
		return 0, 0, fmt.Errorf("invalid number of repository found")
	}
	return repoIDs[0], repoIDs[1], nil
}
