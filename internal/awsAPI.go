package iternal

import (
	"context"
	"fmt"
	"image"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/fogleman/gg"
	"github.com/impopov/aws-recognition-telegram-bot/internal/helpers"
)

func createAWSConfig() (*aws.Config, error) {
	awsKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion := os.Getenv("AWS_REGION")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
		awsKeyID, awsSecretKey, "")))
	if err != nil {
		return nil, err
	}

	cfg.Region = awsRegion

	return &cfg, nil
}

func setContextAndColors(im image.Image) (*gg.Context, error) {
	// Создаем новый контекст для рисования поверх изображения
	dc := gg.NewContextForImage(im)

	// Задаем цвет и размер линии прямоугольника
	dc.SetHexColor("A6FF96")
	dc.SetLineWidth(3)
	err := dc.LoadFontFace("Montserrat-SemiBold.ttf", 20.0)
	if err != nil {
		return dc, err
	}

	return dc, err
}

func recognizeText(config *aws.Config) (*rekognition.DetectTextOutput, error) {
	clientCV := rekognition.NewFromConfig(*config)

	confidenceThreshold := float32(80.0)

	outputCV, err := clientCV.DetectText(context.TODO(), &rekognition.DetectTextInput{
		Image: &types.Image{
			Bytes: helpers.ConvertImgToByte("./input.png"),
		},
		Filters: &types.DetectTextFilters{
			WordFilter: &types.DetectionFilter{
				MinConfidence: &confidenceThreshold,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return outputCV, nil
}

func drawTextBoundingBox(path string, res *rekognition.DetectTextOutput) error {
	img, err := gg.LoadPNG(path)
	if err != nil {
		return err
	}

	dc, err := setContextAndColors(img)
	if err != nil {
		log.Println(err)
	}

	for _, item := range res.TextDetections {
		if item.Type == "WORD" {
			left := float64(*item.Geometry.BoundingBox.Left) * float64(dc.Width())
			top := float64(*item.Geometry.BoundingBox.Top) * float64(dc.Height())
			width := float64(*item.Geometry.BoundingBox.Width) * float64(dc.Width())
			height := float64(*item.Geometry.BoundingBox.Height) * float64(dc.Height())

			// Рисуем bounding box
			dc.DrawRectangle(left, top, width, height)
			dc.Stroke()

			// Добавляем метку (label) рядом с bounding box
			dc.DrawStringAnchored(*item.DetectedText, left+width/2, top-10, 0.5, 0.5)
		}
	}

	// Сохраняем изображение с bounding boxes и метками в файл
	if err = dc.SavePNG("./output.png"); err != nil {
		return err
	}

	return nil
}

func recognizeTextHandler(config *aws.Config) {
	recognizedText, err := recognizeText(config)
	if err != nil {
		log.Println(err)
	}

	//draw bb
	err = drawTextBoundingBox("./input.png", recognizedText)
	if err != nil {
		fmt.Print(err)
	}
}

func recognizeObject(config *aws.Config) (*rekognition.DetectLabelsOutput, error) {
	clientCV := rekognition.NewFromConfig(*config)

	outputCV, err := clientCV.DetectLabels(context.TODO(), &rekognition.DetectLabelsInput{
		Image: &types.Image{
			Bytes: helpers.ConvertImgToByte("./input.png"),
		},
	})
	if err != nil {
		return nil, err
	}

	return outputCV, nil
}

func drawObjectBoundingBox(path string, res *rekognition.DetectLabelsOutput) error {
	img, err := gg.LoadImage(path)
	if err != nil {
		return err
	}

	dc, err := setContextAndColors(img)
	if err != nil {
		log.Println(err)
	}

	var sliceBB []float64

	for _, item := range res.Labels {

		for _, label := range item.Instances {

			//check if instance have bounding box
			if label.BoundingBox.Left != nil {
				left := float64(*label.BoundingBox.Left) * float64(dc.Width())
				top := float64(*label.BoundingBox.Top) * float64(dc.Height())
				width := float64(*label.BoundingBox.Width) * float64(dc.Width())
				height := float64(*label.BoundingBox.Height) * float64(dc.Height())

				// Рисуем bounding box
				dc.DrawRectangle(left, top, width, height)
				dc.Stroke()

				//check if bb already been
				//TODO map
				for _, bb := range sliceBB {
					if left == bb {
						top = top + 15
					}
				}

				sliceBB = append(sliceBB, left)

				// Добавляем метку (label) рядом с bounding box
				dc.DrawStringAnchored(*item.Name, left+width/2, top-10, 0.5, 0.5)
			}

		}
	}

	// Сохраняем изображение с bounding boxes и метками в файл
	if err = dc.SavePNG("./output.png"); err != nil {
		return err
	}

	return nil
}

func recognizeObjectHandler(config *aws.Config) {
	recognizedObject, err := recognizeObject(config)
	if err != nil {
		log.Println(err)
	}

	//draw bb
	err = drawObjectBoundingBox("./input.png", recognizedObject)
	if err != nil {
		fmt.Print(err)
	}
}

func recognizeNudityHandler(config *aws.Config) ([]string, error) {
	clientCV := rekognition.NewFromConfig(*config)

	outputCV, err := clientCV.DetectModerationLabels(context.TODO(), &rekognition.DetectModerationLabelsInput{
		Image: &types.Image{
			Bytes: helpers.ConvertImgToByte("./input.png"),
		},
	})
	if err != nil {
		return nil, err
	}

	var res []string

	for _, item := range outputCV.ModerationLabels {
		res = append(res, *item.Name)
	}

	return res, nil
}

func recognizePPE(config *aws.Config) (*rekognition.DetectProtectiveEquipmentOutput, error) {
	clientCV := rekognition.NewFromConfig(*config)

	//confidenceThreshold := float32(80.0)

	outputCV, err := clientCV.DetectProtectiveEquipment(context.TODO(), &rekognition.DetectProtectiveEquipmentInput{
		Image: &types.Image{
			Bytes: helpers.ConvertImgToByte("./input.png"),
		},
	})
	if err != nil {
		return nil, err
	}

	return outputCV, nil
}

func drawPPEBoundingBox(path string, res *rekognition.DetectProtectiveEquipmentOutput) error {
	img, err := gg.LoadPNG(path)
	if err != nil {
		return err
	}

	dc, err := setContextAndColors(img)
	if err != nil {
		log.Println(err)
	}

	for _, item := range res.Persons {
		left := float64(*item.BoundingBox.Left) * float64(dc.Width())
		top := float64(*item.BoundingBox.Top) * float64(dc.Height())
		width := float64(*item.BoundingBox.Width) * float64(dc.Width())
		height := float64(*item.BoundingBox.Height) * float64(dc.Height())

		// Рисуем bounding box
		dc.DrawRectangle(left, top, width, height)
		dc.Stroke()

		for _, label := range item.BodyParts {
			for _, equipment := range label.EquipmentDetections {
				left = float64(*equipment.BoundingBox.Left) * float64(dc.Width())
				top = float64(*equipment.BoundingBox.Top) * float64(dc.Height())
				width = float64(*equipment.BoundingBox.Width) * float64(dc.Width())
				height = float64(*equipment.BoundingBox.Height) * float64(dc.Height())

				dc.DrawRectangle(left, top, width, height)
				dc.Stroke()
			}

			// Добавляем метку (label) рядом с bounding box
			dc.DrawStringAnchored(string(label.Name), left+width/2, top-10, 0.5, 0.5)

		}

	}

	// Сохраняем изображение с bounding boxes и метками в файл
	if err = dc.SavePNG("./output.png"); err != nil {
		return err
	}

	return nil
}

func recognizePPEHandler(config *aws.Config) {
	recognizedPPE, err := recognizePPE(config)
	if err != nil {
		log.Println(err)
	}

	//draw bb
	err = drawPPEBoundingBox("./input.png", recognizedPPE)
	if err != nil {
		fmt.Print(err)
	}
}
