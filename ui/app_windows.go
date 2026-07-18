//go:build windows && cgo

package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"AVITOUO/core"
	"AVITOUO/storage"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyneApp   fyne.App
	window    fyne.Window
	settings  *core.Settings
	excelFile *excelizeStub
	currentAd *core.AdRow
}

type excelizeStub struct {
	path string
}

func NewApp() *App {
	a := app.New()
	w := a.NewWindow("Редактор шаблонов Авито")
	w.Resize(fyne.NewSize(1000, 700))

	settings, _ := storage.LoadSettings()

	return &App{
		fyneApp:  a,
		window:   w,
		settings: settings,
	}
}

func (app *App) Run() {
	tabs := container.NewAppTabs(
		container.NewTabItem("Инструкция", app.buildInstructionTab()),
		container.NewTabItem("Настройки", app.buildSettingsTab()),
		container.NewTabItem("Редактор", app.buildEditorTab()),
		container.NewTabItem("Фото и Экспорт", app.buildExportTab()),
	)

	app.window.SetContent(tabs)
	app.window.ShowAndRun()
}

func (app *App) buildInstructionTab() fyne.CanvasObject {
	text := `Как использовать шаблон:
1. Лист "Инструкция" переименовывать нельзя.
2. В листах категорий строки 1–4 защищены от удаления, изменения и смены порядка.
3. Заполнение данных начинается строго с 5-й строки.
4. Каждое объявление в отдельной строке, объединение ячеек запрещено.
5. Лимит: не более 50 000 объявлений в файле.
6. Уникальный идентификатор генерируется автоматически.
7. Описание автоматически оборачивается в <![CDATA[ ... ]]>, переносы строк заменяются на <br>.`
	return widget.NewLabel(text)
}

func (app *App) buildSettingsTab() fyne.CanvasObject {
	contactsEntry := widget.NewMultiLineEntry()
	contactsEntry.SetText(strings.Join(app.settings.Contacts, "\n"))

	saveBtn := widget.NewButton("Сохранить настройки", func() {
		app.settings.Contacts = strings.Split(strings.TrimSpace(contactsEntry.Text()), "\n")
		storage.SaveSettings(app.settings)
		dialog.ShowInformation("Успех", "Настройки сохранены", app.window)
	})

	return container.NewVBox(
		widget.NewLabel("Контактные лица (каждый с новой строки):"),
		contactsEntry,
		saveBtn,
	)
}

func (app *App) buildEditorTab() fyne.CanvasObject {
	loadBtn := widget.NewButton("Загрузить шаблон XLSX", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			dialog.ShowInformation("Загрузка", fmt.Sprintf("Загружен: %s", reader.URI().Name()), app.window)
		}, app.window)
	})

	adID := widget.NewEntry()
	adID.SetText(core.GenerateUniqueID())
	adID.Disable()

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Введите название (до 100 символов)")

	shuffleBtn := widget.NewButton("Перемешать слова", func() {
		dialog.ShowInformation("Инфо", "Функция перемешивания слов", app.window)
	})

	validateBtn := widget.NewButton("Проверить валидацию", func() {
		if err := core.ValidateTitle(titleEntry.Text); err != nil {
			dialog.ShowError(err, app.window)
		} else {
			dialog.ShowInformation("OK", "Название корректно", app.window)
		}
	})

	return container.NewVBox(
		loadBtn,
		widget.NewForm(
			widget.NewFormItem("Уникальный ID", adID),
			widget.NewFormItem("Название объявления", titleEntry),
		),
		container.NewHBox(shuffleBtn, validateBtn),
	)
}

func (app *App) buildExportTab() fyne.CanvasObject {
	exportBtn := widget.NewButton("Создать ZIP и Сохранить XLSX", func() {
		dialog.ShowConfirm("Экспорт", "Начать процесс экспорта и проверки лимита 100 МБ?", func(ok bool) {
			if ok {
				err := storage.CheckTotalSize("dummy.zip", "dummy.xlsx")
				if err != nil {
					dialog.ShowError(err, app.window)
				} else {
					dialog.ShowInformation("Успех", "Файлы готовы к выгрузке", app.window)
				}
			}
		}, app.window)
	})

	return container.NewCenter(exportBtn)
}
