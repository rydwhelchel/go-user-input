package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

type Album struct {
    ID     int64
    Title  string
    Artist string
    Price  float32
}

var db *sql.DB

func main() {
    // Capture connection properties.
    cfg := mysql.Config{
        User:   os.Getenv("DBUSER"),
        Passwd: os.Getenv("DBPASS"),
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        DBName: "go_test",
    }

    // Get a database handle.
    var err error
    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
        log.Fatal(err)
    }

    pingErr := db.Ping()
    if pingErr != nil {
        log.Fatal(pingErr)
    }
    fmt.Println("Connected!")

    albums, err := albumsByArtist("John Coltrane")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Albums: %v\n", albums)

    // Input start
    var command string
    fmt.Println("Please enter a command (search, get, add):")

    inputReader := bufio.NewReader(os.Stdin)
    // Stop parsing input when a newline is inserted
    command, _ = inputReader.ReadString('\n')

    direction := strings.Split(strings.TrimSuffix(command, "\n"), " ")
    output,err := processDirection(direction)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("You do love typing: %v\n", output)
}

func processDirection(dir []string) (string, error) {
    switch dir[0] {
    case "get":
        id,err := strconv.Atoi(dir[1])
        if err != nil {
            return "",fmt.Errorf("Error processing get direction ID: %v", dir[1])
        }
        getResponse,err := albumByID(int64(id))
        return fmt.Sprintf("Successfully retrieved album with ID %v: %v", id, getResponse),nil
    case "search":
        name := buildName(dir)
        searchResponse,err := albumsByArtist(name)
        if err != nil {
            return "",err
        }
        return fmt.Sprintf("Successfully found albums for artist %v: %v", name, searchResponse),nil
    case "add":
        response,err := promptNewAlbum()
        if err != nil {
            return "",err
        }
        id,err := addAlbum(response)
        if err != nil {
            return "",err
        }
        return fmt.Sprintf("Successfully added album, id is %v", id),nil
    default:
        return "",fmt.Errorf("Error processing dir: %v", dir)
    }
}

func buildName(dir []string) string {
    builder := strings.Builder{}
    for i,v := range dir {
        if i == 0 {
            continue;
        }
        fmt.Fprintf(&builder, "%v ", v)
    }
    output := builder.String()
    return strings.TrimSuffix(output, " ")
}

func promptNewAlbum() (Album, error) {
    album := Album{}
    inputReader := bufio.NewReader(os.Stdin)
    fmt.Print("Please enter the albums title: ")
    title,err := inputReader.ReadString('\n')
    title = strings.TrimSuffix(title, "\n")
    if err != nil {
        return album,fmt.Errorf("Error processing the albums title")
    }
    fmt.Print("Please enter the albums artist: ")
    artist,err := inputReader.ReadString('\n')
    artist = strings.TrimSuffix(artist, "\n")
    if err != nil {
        return album,fmt.Errorf("Error processing the albums artist")
    }
    fmt.Print("Please enter the albums price: ")
    price,err := inputReader.ReadString('\n')
    price = strings.TrimSuffix(price, "\n")
    if err != nil {
        return album,fmt.Errorf("Error processing the albums price")
    }
    priceVal,err := strconv.ParseFloat(price, 32)
    if err != nil {
        return album,fmt.Errorf("Error converting the albums price: %v\n, %v", price, err)
    }
    album.Title = title
    album.Artist = artist
    album.Price = float32(priceVal)
    return album,nil
}

// albumsByArtist queries for albums that have the specified artist name.
func albumsByArtist(name string) ([]Album, error) {
    fmt.Printf("Getting albums by %v...\n", name)
    // An albums slice to hold data from returned rows.
    var albums []Album

    rows, err := db.Query("SELECT * FROM album WHERE artist = ?", name)
    if err != nil {
        return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
    }
    defer rows.Close()
    // Loop through rows, using Scan to assign column data to struct fields.
    for rows.Next() {
        var alb Album
        if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
            return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
        }
        albums = append(albums, alb)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
    }
    return albums, nil
}

// albumByID queries for the album with the specified ID.
func albumByID(id int64) (Album, error) {
    fmt.Printf("Getting album with id %v...\n", id)
    // An album to hold data from the returned row.
    var alb Album

    row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)
    if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
        if err == sql.ErrNoRows {
            return alb, fmt.Errorf("albumsById %d: no such album", id)
        }
        return alb, fmt.Errorf("albumsById %d: %v", id, err)
    }
    return alb, nil
}

// addAlbum adds the specified album to the database,
// returning the album ID of the new entry
func addAlbum(alb Album) (int64, error) {
    fmt.Printf("Inserting album: title: %v, artist: %v, price: %v\n", alb.Title, alb.Artist, alb.Price)
    result, err := db.Exec("INSERT INTO album (title, artist, price) VALUES (?, ?, ?)", alb.Title, alb.Artist, alb.Price)
    if err != nil {
        return 0, fmt.Errorf("addAlbum: %v", err)
    }
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("addAlbum: %v", err)
    }
    return id, nil
}
