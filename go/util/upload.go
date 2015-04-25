package util

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"image"
	"image/png"
	_ "image/jpeg"
	_ "image/gif"
	"github.com/disintegration/imaging"
	"mime/multipart"
)

func SaveImage(db *sql.DB, user string, file multipart.File, header *multipart.FileHeader) error {
	picture, _, err := image.Decode(file)
	if err != nil {
		return NewError(nil, "Formato de imagen invalido", 400)
	}

	bounds := picture.Bounds()
	ratio := float64(bounds.Dx())/float64(bounds.Dy())
	/*
	if ratio < .8 || ratio > 1.2 {
		return NewError(nil, "Dimensiones de imagen inválidos", 400)
	}
	*/

	cropWidth := bounds.Dx();

	if ratio > 1 {
		cropWidth = bounds.Dy()
	}
	croppedPicture := imaging.CropCenter(picture, cropWidth, cropWidth)

	pictureNormal := imaging.Resize(croppedPicture, 100, 100, imaging.Lanczos)
	pictureSmall := imaging.Resize(croppedPicture, 50, 50, imaging.Lanczos)
	pictureThumbnail := imaging.Resize(croppedPicture, 35, 35, imaging.Lanczos)

	img, _ := os.Create("/var/www/html/images/" + user + ".png")
	defer img.Close()
	err = png.Encode(img, pictureNormal)	
	if err != nil {
		return NewError(err, "Imagen no puede ser usada", 500)
	}

	imgSmall, _ := os.Create("/var/www/html/images/" + user + "_50.png")
	defer imgSmall.Close()
	err = png.Encode(imgSmall, pictureSmall)	
	if err != nil {
		return NewError(err, "Imagen no puede ser usada", 500)
	}

	imgThumnnail, _ := os.Create("/var/www/html/images/" + user + "_35.png")
	defer imgThumnnail.Close()
	err = png.Encode(imgThumnnail, pictureThumbnail)	
	if err != nil {
		return NewError(err, "Imagen no puede ser usada", 500)
	}

	err = setCustomPicture(db, true, user)
	if err != nil { return err }

	return nil
}

func setCustomPicture(db *sql.DB, customPicture bool, user string) error {
	stmt, err := db.Prepare(`
		UPDATE users
			SET custom_picture = ?
			WHERE users.name = ?;
		`)
	if err != nil {
		return NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(customPicture, user)
	if err != nil {
		return NewError(err, "Error de la base de datos", 500)
	}
	return nil
}

func DeletePicture(db *sql.DB, user string) error {
	err := setCustomPicture(db, false, user)
	if err != nil { return err }

	err = os.Remove("/var/www/html/images/" + user + ".png")
	if err != nil { return NewError(err, "Error borrando imagen", 500) }
	err = os.Remove("/var/www/html/images/" + user + "_50.png")
	if err != nil { return NewError(err, "Error borrando imagen", 500) }
	err = os.Remove("/var/www/html/images/" + user + "_35.png")
	if err != nil { return NewError(err, "rror borrando imagen", 500) }

	return nil
}