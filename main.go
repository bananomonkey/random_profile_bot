package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "modernc.org/sqlite"
)

const (
	styleReal    = "real"
	stylePixel   = "pixel"
	styleCartoon = "cartoon"

	fakeNamelyURL = "https://fakenamely.com/api/v1/identity"
	randomUserURL = "https://randomuser.me/api/"
)

type Profile struct {
	FirstName  string
	LastName   string
	Patronymic string
	Gender     string
	Callsign   string
	Blood      string
	BirthDate  time.Time
	Age        int
	City       string
	Street     string
	House      int
	Phone      string
	Username   string
	Passport   string
	AvatarURL  string
}

type fakeNamelyResponse struct {
	Success bool         `json:"success"`
	Data    []fakePerson `json:"data"`
}

type fakePerson struct {
	Name struct {
		Full      string `json:"full"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Middle    string `json:"middle"`
	} `json:"name"`
	Phone    string `json:"phone"`
	Username string `json:"username"`
	Sex      string `json:"sex"`
	Personal struct {
		Birthday string `json:"birthday"`
		Age      int    `json:"age"`
	} `json:"personal"`
}

type randomUserResponse struct {
	Results []struct {
		Picture struct {
			Large string `json:"large"`
		} `json:"picture"`
	} `json:"results"`
}

var nonUsername = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

var (
	cities = []string{
		"Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань",
		"Нижний Новгород", "Челябинск", "Самара", "Омск", "Ростов-на-Дону",
		"Уфа", "Красноярск", "Пермь", "Воронеж", "Волгоград", "Краснодар",
		"Саратов", "Тюмень", "Тольятти", "Ижевск", "Барнаул", "Ульяновск",
		"Иркутск", "Хабаровск", "Ярославль", "Владивосток", "Махачкала", "Томск",
		"Оренбург", "Кемерово", "Новокузнецк", "Рязань", "Астрахань", "Пенза",
		"Липецк", "Киров", "Чебоксары", "Тула", "Калининград", "Курск",
		"Сочи", "Ставрополь", "Белгород", "Брянск", "Владимир", "Архангельск",
		"Сургут", "Якутск", "Мурманск", "Грозный",
	}

	streets = []string{
		"Ленинградская", "Советская", "Молодёжная", "Центральная", "Победы",
		"Космонавтов", "Октябрьская", "Комсомольская", "Садовая", "Школьная",
		"Новая", "Речная", "Заречная", "Полевая", "Лесная", "Пролетарская",
		"Кирова", "Гагарина", "Мира", "Пушкина", "Лермонтова", "Горького",
		"Толстого", "Чехова", "Тургенева", "Жукова", "Рокоссовского", "Суворова",
		"Кутузова", "Нахимова", "Матросова", "Чкалова", "Королёва", "Циолковского",
		"Бауманская", "Тверская", "Арбатская", "Кузнецкая", "Севастопольская",
		"Парковая", "Береговая", "Набережная", "Северная", "Южная", "Восточная",
		"Западная", "Высотная", "Спортивная", "Тихая", "Транспортная", "Энергетиков",
		"Строителей", "Машиностроителей", "Заводская", "Флотская", "Кедровая", "Таёжная",
		"Сосновая", "Рябиновая", "Вишнёвая", "Ягодная", "Кленовая", "Берёзовая",
		"Тихоокеанская", "Сибирская", "Уральская", "Донская", "Волжская", "Кавказская",
		"Байкальская", "Балтийская", "Карельская", "Дальневосточная", "Крымская", "Каспийская",
		"Московская", "Петербургская", "Ростовская", "Казанская", "Самарская", "Тюменская",
		"Красноярская", "Омская", "Воронежская", "Саратовская", "Уфимская", "Пермская",
		"Техническая", "Индустриальная", "Академическая", "Университетская", "Пионерская",
		"Магистральная", "Шоссейная", "Кольцевая", "Дорожная", "Политехническая", "Рабочая",
		"Фронтовая", "Героев", "Сиреневая", "Каштановая", "Кленовая", "Луговая",
		"Озерная", "Горная", "Высокая", "Верхняя", "Нижняя", "Каменская", "Станционная",
	}

	maleNames = []string{
		"Александр", "Алексей", "Андрей", "Антон", "Аркадий", "Артём", "Артур", "Богдан",
		"Борис", "Вадим", "Валентин", "Валерий", "Василий", "Виктор", "Виталий", "Владимир",
		"Владислав", "Вячеслав", "Геннадий", "Георгий", "Глеб", "Григорий", "Данил", "Даниил",
		"Денис", "Дмитрий", "Евгений", "Егор", "Елисей", "Захар", "Иван", "Игнат", "Игорь",
		"Илья", "Иннокентий", "Иосиф", "Кирилл", "Константин", "Лев", "Леонид", "Макар",
		"Максим", "Марк", "Матвей", "Михаил", "Назар", "Никита", "Николай", "Олег", "Павел",
		"Пётр", "Платон", "Ратмир", "Ринат", "Роман", "Ростислав", "Руслан", "Савелий",
		"Святослав", "Семён", "Сергей", "Степан", "Тарас", "Тимофей", "Тимур", "Тихон",
		"Фёдор", "Филипп", "Эдуард", "Эльдар", "Юрий", "Ярослав", "Демид", "Арсений",
		"Виссарион", "Викентий", "Всеволод", "Гордей", "Давид", "Дорофей", "Клим", "Лука",
		"Мирон", "Прохор", "Родион", "Рустам", "Трофим", "Шамиль", "Ян", "Ефим", "Ираклий",
		"Альберт", "Вениамин", "Мстислав", "Остап", "Святогор", "Фадей", "Юлиан", "Рудольф",
	}

	femaleNames = []string{
		"Алина", "Алёна", "Анастасия", "Анжела", "Анна", "Валерия", "Варвара", "Василиса",
		"Вера", "Вероника", "Виктория", "Владислава", "Галина", "Дарья", "Диана", "Ева",
		"Евгения", "Екатерина", "Елена", "Елизавета", "Жанна", "Злата", "Зинаида", "Ирина",
		"Кира", "Кристина", "Ксения", "Лариса", "Лидия", "Любовь", "Людмила", "Маргарита",
		"Марина", "Мария", "Милана", "Мирослава", "Надежда", "Наталья", "Нелли", "Ника",
		"Оксана", "Олеся", "Ольга", "Полина", "Раиса", "Регина", "Римма", "Светлана",
		"София", "Снежана", "Таисия", "Тамара", "Татьяна", "Ульяна", "Фаина", "Юлия",
		"Яна", "Ярослава", "Агата", "Аделина", "Арина", "Богдана", "Василина", "Веста",
		"Глафира", "Доминика", "Есения", "Зоряна", "Инна", "Карина", "Клавдия", "Лилия",
		"Майя", "Мелания", "Нина", "Олеся", "Прасковья", "Рада", "Серафима", "Стефания",
		"Элина", "Эмилия", "Эльвира", "Аксинья", "Алевтина", "Белла", "Виолетта", "Дина",
		"Зоя", "Лиана", "Луиза", "Марта", "Розалия", "Сусанна", "Эвелина", "Элеонора",
	}

	surnames = []string{
		"Абрамов", "Агапов", "Акимов", "Аксенов", "Александров", "Алексеев", "Антонов", "Артемьев",
		"Баранов", "Беляев", "Белов", "Белоусов", "Березин", "Блинов", "Бобров", "Богданов",
		"Большаков", "Борисов", "Брагин", "Быков", "Васильев", "Виноградов", "Власов", "Волков",
		"Воронин", "Гаврилов", "Гайдук", "Галкин", "Герасимов", "Гладков", "Голубев", "Гончаров",
		"Горбунов", "Гордеев", "Грачев", "Гришин", "Громов", "Давыдов", "Данилов", "Демидов",
		"Денисов", "Дмитриев", "Добрынин", "Дроздов", "Егоров", "Елисеев", "Ершов", "Жданов",
		"Жуков", "Завьялов", "Зайцев", "Захаров", "Зимин", "Зорин", "Иванов", "Игнатов",
		"Ильин", "Калинин", "Каменев", "Капустин", "Карпов", "Касаткин", "Киселев", "Климов",
		"Князев", "Ковалев", "Колесников", "Комаров", "Кондратьев", "Королев", "Котов", "Кочетков",
		"Кравцов", "Крылов", "Крюков", "Кудрявцев", "Кузнецов", "Куликов", "Курочкин", "Лазарев",
		"Лапин", "Ларионов", "Лебедев", "Леонов", "Лихачев", "Лобанов", "Логинов", "Лукин",
		"Макаров", "Максимов", "Малинин", "Малышев", "Мартынов", "Маслов", "Матвеев", "Медведев",
		"Мельников", "Меркулов", "Миронов", "Михайлов", "Молчанов", "Морозов", "Моисеев", "Муравьев",
		"Назаров", "Наумов", "Нестеров", "Нечаев", "Никитин", "Николаев", "Новиков", "Овсянников",
		"Орлов", "Осипов", "Павлов", "Панов", "Пархоменко", "Петров", "Платонов", "Поляков",
		"Пономарев", "Попов", "Потапов", "Прохоров", "Рогов", "Родионов", "Романов", "Руденко",
		"Рыбаков", "Рябов", "Савельев", "Самсонов", "Сафонов", "Сахаров", "Селезнев", "Семенов",
		"Сергиев", "Сидоров", "Симонов", "Смирнов", "Соболев", "Соколов", "Соловьев", "Сорокин",
		"Степанов", "Стрелков", "Субботин", "Суворов", "Тарасов", "Терентьев", "Тихонов", "Токарев",
		"Третьяков", "Трофимов", "Уваров", "Устинов", "Федоров", "Филиппов", "Фомин", "Фролов",
		"Харитонов", "Хохлов", "Цветков", "Чернов", "Чернышев", "Чистяков", "Шаповалов", "Шарапов",
		"Шевцов", "Шестаков", "Широков", "Шубин", "Щербаков", "Юдин", "Юрьев", "Яковлев",
		"Якубов", "Алёхин", "Балашов", "Барсуков", "Бессонов", "Буров", "Веденин", "Вершинин",
		"Доронин", "Кудин", "Левин", "Минин", "Островский", "Панфилов", "Ратников", "Сазонов",
		"Туманов", "Фокин", "Шувалов", "Гусев", "Кузьмин", "Ларионов", "Кулиш", "Мазур",
		"Савин", "Сафронов", "Лыткин", "Мороз", "Скориков", "Белкин", "Веденин", "Серов",
	}

	callsigns = []string{
		"Гранит", "Седой", "Ворон", "Чекист", "Алтай", "Скиф", "Шаман", "Туман", "Варяг", "Компас",
		"Струна", "Лютый", "Кремень", "Гром", "Борз", "Маэстро", "Барс", "Тайфун", "Шторм", "Буран",
		"Ветер", "Север", "Юг", "Восток", "Запад", "Феникс", "Сокол", "Ястреб", "Коршун", "Орёл",
		"Рысь", "Тигр", "Волк", "Медведь", "Лось", "Бык", "Кабан", "Кедр", "Клён", "Тополь",
		"Сапсан", "Кобра", "Гюрза", "Удав", "Шершень", "Шмель", "Овод", "Беркут", "Кондор", "Клык",
		"Коготь", "Бритва", "Молот", "Кувалда", "Нож", "Кремень", "Искра", "Пламя", "Уголь", "Дым",
		"Гарпун", "Таран", "Резак", "Стилет", "Кинжал", "Пуля", "Залп", "Калибр", "Контур", "Вектор",
		"Радар", "Сигнал", "Пеленг", "Факел", "Маяк", "Форт", "Бастион", "Редут", "Барьер", "Щит",
		"Дозор", "Караул", "Скала", "Утёс", "Вершина", "Каньон", "Пик", "Каскад", "Поток", "Прибой",
		"Риф", "Шельф", "Океан", "Дельта", "Вихрь", "Циклон", "Гроза", "Молния", "Смерч", "Зенит",
		"Орион", "Кедр", "Тайга", "Полюс", "Ладога", "Байкал", "Енисей", "Дон", "Волга", "Амур",
		"Дунай", "Талисман", "Карат", "Самородок", "Рубеж", "Курс", "Маркер", "Шифр", "Код", "Призрак",
		"Тень", "Мираж", "Фантом", "Сфинкс", "Аргумент", "Форсаж", "Драйв", "Титан", "Атлас", "Прометей",
	}

	bloodTypes = []string{
		"I (O) Rh+", "I (O) Rh-", "II (A) Rh+", "II (A) Rh-",
		"III (B) Rh+", "III (B) Rh-", "IV (AB) Rh+", "IV (AB) Rh-",
	}

	usernameWords = []string{
		"granite", "sedoy", "voron", "skif", "shaman", "tuman", "varyag", "kompas", "kremen", "grom",
		"bars", "tayfun", "shtorm", "buran", "sokol", "yastreb", "klyak", "molot", "iskra", "druzhina",
		"sever", "vektor", "radar", "signal", "fort", "redut", "dozor", "skala", "zenit", "orion",
		"baykal", "amur", "rubey", "marker", "shifr", "titan", "atlas", "fantom", "ten", "prizrak",
	}

	fatherPatronymics = map[string][2]string{
		"Александр": {"Александрович", "Александровна"}, "Алексей": {"Алексеевич", "Алексеевна"},
		"Андрей": {"Андреевич", "Андреевна"}, "Антон": {"Антонович", "Антоновна"},
		"Аркадий": {"Аркадьевич", "Аркадьевна"}, "Артём": {"Артёмович", "Артёмовна"},
		"Артур": {"Артурович", "Артуровна"}, "Богдан": {"Богданович", "Богдановна"},
		"Борис": {"Борисович", "Борисовна"}, "Вадим": {"Вадимович", "Вадимовна"},
		"Валентин": {"Валентинович", "Валентиновна"}, "Валерий": {"Валерьевич", "Валерьевна"},
		"Василий": {"Васильевич", "Васильевна"}, "Виктор": {"Викторович", "Викторовна"},
		"Виталий": {"Витальевич", "Витальевна"}, "Владимир": {"Владимирович", "Владимировна"},
		"Владислав": {"Владиславович", "Владиславовна"}, "Вячеслав": {"Вячеславович", "Вячеславовна"},
		"Геннадий": {"Геннадьевич", "Геннадьевна"}, "Георгий": {"Георгиевич", "Георгиевна"},
		"Глеб": {"Глебович", "Глебовна"}, "Григорий": {"Григорьевич", "Григорьевна"},
		"Данил": {"Данилович", "Даниловна"}, "Даниил": {"Данилович", "Даниловна"},
		"Денис": {"Денисович", "Денисовна"}, "Дмитрий": {"Дмитриевич", "Дмитриевна"},
		"Евгений": {"Евгеньевич", "Евгеньевна"}, "Егор": {"Егорович", "Егоровна"},
		"Елисей": {"Елисеевич", "Елисеевна"}, "Захар": {"Захарович", "Захаровна"},
		"Иван": {"Иванович", "Ивановна"}, "Игнат": {"Игнатович", "Игнатовна"},
		"Игорь": {"Игоревич", "Игоревна"}, "Илья": {"Ильич", "Ильинична"},
		"Иннокентий": {"Иннокентьевич", "Иннокентьевна"}, "Иосиф": {"Иосифович", "Иосифовна"},
		"Кирилл": {"Кириллович", "Кирилловна"}, "Константин": {"Константинович", "Константиновна"},
		"Лев": {"Львович", "Львовна"}, "Леонид": {"Леонидович", "Леонидовна"},
		"Макар": {"Макарович", "Макаровна"}, "Максим": {"Максимович", "Максимовна"},
		"Марк": {"Маркович", "Марковна"}, "Матвей": {"Матвеевич", "Матвеевна"},
		"Михаил": {"Михайлович", "Михайловна"}, "Назар": {"Назарович", "Назаровна"},
		"Никита": {"Никитич", "Никитична"}, "Николай": {"Николаевич", "Николаевна"},
		"Олег": {"Олегович", "Олеговна"}, "Павел": {"Павлович", "Павловна"},
		"Пётр": {"Петрович", "Петровна"}, "Платон": {"Платонович", "Платоновна"},
		"Роман": {"Романович", "Романовна"}, "Ростислав": {"Ростиславович", "Ростиславовна"},
		"Руслан": {"Русланович", "Руслановна"}, "Савелий": {"Савельевич", "Савельевна"},
		"Святослав": {"Святославович", "Святославовна"}, "Семён": {"Семёнович", "Семёновна"},
		"Сергей": {"Сергеевич", "Сергеевна"}, "Степан": {"Степанович", "Степановна"},
		"Тарас": {"Тарасович", "Тарасовна"}, "Тимофей": {"Тимофеевич", "Тимофеевна"},
		"Тимур": {"Тимурович", "Тимуровна"}, "Тихон": {"Тихонович", "Тихоновна"},
		"Фёдор": {"Фёдорович", "Фёдоровна"}, "Филипп": {"Филиппович", "Филипповна"},
		"Эдуард": {"Эдуардович", "Эдуардовна"}, "Юрий": {"Юрьевич", "Юрьевна"},
		"Ярослав": {"Ярославович", "Ярославовна"}, "Арсений": {"Арсеньевич", "Арсеньевна"},
		"Всеволод": {"Всеволодович", "Всеволодовна"}, "Давид": {"Давидович", "Давидовна"},
		"Клим": {"Климович", "Климовна"}, "Лука": {"Лукич", "Лукинична"},
		"Мирон": {"Миронович", "Мироновна"}, "Прохор": {"Прохорович", "Прохоровна"},
		"Родион": {"Родионович", "Родионовна"}, "Рустам": {"Рустамович", "Рустамовна"},
		"Трофим": {"Трофимович", "Трофимовна"}, "Шамиль": {"Шамильевич", "Шамильевна"},
		"Ян": {"Янович", "Яновна"}, "Ефим": {"Ефимович", "Ефимовна"},
		"Альберт": {"Альбертович", "Альбертовна"}, "Вениамин": {"Вениаминович", "Вениаминовна"},
		"Мстислав": {"Мстиславович", "Мстиславовна"}, "Остап": {"Остапович", "Остаповна"},
		"Фадей": {"Фадеевич", "Фадеевна"}, "Рудольф": {"Рудольфович", "Рудольфовна"},
	}
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	token := strings.TrimSpace(os.Getenv("BOT_TOKEN"))
	if token == "" {
		log.Fatal("BOT_TOKEN is not set")
	}

	db, err := openDB("profiles.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("authorized as @%s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.CallbackQuery != nil {
			handleCallback(bot, db, update.CallbackQuery)
			continue
		}
		if update.Message == nil {
			continue
		}
		handleMessage(bot, db, update.Message)
	}
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Для маленького Telegram-бота одного соединения достаточно и уменьшает шанс SQLITE_BUSY.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA busy_timeout = 5000;",
	}
	for _, q := range pragmas {
		if _, err := db.Exec(q); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("sqlite pragma: %w", err)
		}
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_settings (
			user_id INTEGER PRIMARY KEY,
			avatar_style TEXT NOT NULL DEFAULT 'real',
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		);
	`)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}
	return db, nil
}

func handleMessage(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	if !message.IsCommand() {
		if _, err := bot.Send(menuMessage(message.Chat.ID)); err != nil {
			log.Printf("send menu: %v", err)
		}
		return
	}

	switch strings.ToLower(message.Command()) {
	case "start", "help":
		if _, err := bot.Send(menuMessage(message.Chat.ID)); err != nil {
			log.Printf("send start menu: %v", err)
		}
	default:
		if _, err := bot.Send(menuMessage(message.Chat.ID)); err != nil {
			log.Printf("send default menu: %v", err)
		}
	}
}

func handleCallback(bot *tgbotapi.BotAPI, db *sql.DB, callback *tgbotapi.CallbackQuery) {
	answer := tgbotapi.NewCallback(callback.ID, "")
	if _, err := bot.Request(answer); err != nil {
		log.Printf("answer callback: %v", err)
	}

	chatID := callback.Message.Chat.ID
	userID := callback.From.ID

	switch callback.Data {
	case "generate_profile":
		generateAndSendProfile(bot, db, chatID, userID)
	case "settings":
		style, err := getUserStyle(db, userID)
		if err != nil {
			log.Printf("get style: %v", err)
			return
		}
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			chatID,
			callback.Message.MessageID,
			settingsText(style),
			settingsKeyboard(style),
		)
		edit.ParseMode = tgbotapi.ModeMarkdownV2
		if _, err := bot.Send(edit); err != nil {
			log.Printf("edit settings: %v", err)
		}
	case "style_real", "style_pixel", "style_cartoon":
		style := styleReal
		switch callback.Data {
		case "style_pixel":
			style = stylePixel
		case "style_cartoon":
			style = styleCartoon
		}
		if err := setUserStyle(db, userID, style); err != nil {
			log.Printf("set style: %v", err)
			return
		}
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			chatID,
			callback.Message.MessageID,
			settingsText(style),
			settingsKeyboard(style),
		)
		edit.ParseMode = tgbotapi.ModeMarkdownV2
		if _, err := bot.Send(edit); err != nil {
			log.Printf("edit style menu: %v", err)
		}
	case "back":
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			chatID,
			callback.Message.MessageID,
			"🪖 *Генератор профилей*\n\nНажми кнопку ниже — бот создаст новый полностью вымышленный профиль\\.",
			mainKeyboard(),
		)
		edit.ParseMode = tgbotapi.ModeMarkdownV2
		if _, err := bot.Send(edit); err != nil {
			log.Printf("edit back: %v", err)
		}
	}
}

func generateAndSendProfile(bot *tgbotapi.BotAPI, db *sql.DB, chatID, userID int64) {
	style, err := getUserStyle(db, userID)
	if err != nil {
		log.Printf("get style before generation: %v", err)
		style = styleReal
	}

	profile := generateProfile(style)
	caption := formatProfile(profile)

	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(profile.AvatarURL))
	photo.Caption = caption
	photo.ParseMode = tgbotapi.ModeMarkdownV2
	photo.ReplyMarkup = mainKeyboard()

	if _, err := bot.Send(photo); err != nil {
		log.Printf("send profile photo: %v", err)
		// Резерв: даже если Telegram не смог забрать изображение по URL, досье не теряется.
		msg := tgbotapi.NewMessage(chatID, caption)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = mainKeyboard()
		if _, sendErr := bot.Send(msg); sendErr != nil {
			log.Printf("send profile text fallback: %v", sendErr)
		}
	}
}

func menuMessage(chatID int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, "🪖 *Генератор профилей*\n\nНажми кнопку ниже — бот создаст новый полностью вымышленный профиль\\.")
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = mainKeyboard()
	return msg
}

func mainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🪪 Сгенерировать профиль", "generate_profile"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Настройки", "settings"),
		),
	)
}

func settingsKeyboard(style string) tgbotapi.InlineKeyboardMarkup {
	button := func(text, data, currentStyle string) tgbotapi.InlineKeyboardButton {
		if currentStyle == style {
			text = "✅ " + text
		}
		return tgbotapi.NewInlineKeyboardButtonData(text, data)
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			button("Реальное лицо", "style_real", styleReal),
		),
		tgbotapi.NewInlineKeyboardRow(
			button("Пиксельный аватар", "style_pixel", stylePixel),
		),
		tgbotapi.NewInlineKeyboardRow(
			button("Мультяшный / Векторный", "style_cartoon", styleCartoon),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back"),
		),
	)
}

func settingsText(style string) string {
	current := map[string]string{
		styleReal:    "📷 Реальное лицо",
		stylePixel:   "👾 Пиксельный аватар",
		styleCartoon: "🎨 Мультяшный / Векторный",
	}[style]
	if current == "" {
		current = "👾 Пиксельный аватар"
	}
	return fmt.Sprintf("⚙️ *Настройки аватара*\n\nТекущий стиль: *%s*\n\nВыбери стиль для следующей генерации\\.", escapeMD(current))
}

func getUserStyle(db *sql.DB, userID int64) (string, error) {
	var style string
	err := db.QueryRow("SELECT avatar_style FROM user_settings WHERE user_id = ?", userID).Scan(&style)
	if err == sql.ErrNoRows {
		style = stylePixel
		_, err = db.Exec("INSERT INTO user_settings (user_id, avatar_style, updated_at) VALUES (?, ?, strftime('%s','now'))", userID, style)
		return style, err
	}
	return style, err
}

func setUserStyle(db *sql.DB, userID int64, style string) error {
	if style != styleReal && style != stylePixel && style != styleCartoon {
		style = stylePixel
	}
	_, err := db.Exec(`
		INSERT INTO user_settings (user_id, avatar_style, updated_at)
		VALUES (?, ?, strftime('%s','now'))
		ON CONFLICT(user_id) DO UPDATE SET
			avatar_style = excluded.avatar_style,
			updated_at = strftime('%s','now')
	`, userID, style)
	return err
}

func generateProfile(style string) Profile {
	genderAPI := "male"
	if rand.Intn(2) == 1 {
		genderAPI = "female"
	}

	if person, err := fetchFakePerson(genderAPI); err == nil {
		firstName := strings.TrimSpace(person.Name.FirstName)
		lastName := strings.TrimSpace(person.Name.LastName)
		if firstName == "" || lastName == "" {
			parts := strings.Fields(person.Name.Full)
			if len(parts) >= 2 {
				lastName, firstName = parts[0], parts[1]
			}
		}

		if firstName != "" && lastName != "" {
			male := strings.EqualFold(genderAPI, "male")
			gender := "Мужской"
			if !male {
				gender = "Женский"
				lastName = feminineSurname(lastName)
			}

			birthDate, age := apiBirthDate(person.Personal.Birthday)
			if birthDate.IsZero() || age < 20 || age > 50 {
				birthDate, age = randomBirthDate()
			}

			patronymic := strings.TrimSpace(person.Name.Middle)
			if patronymic == "" {
				patronymic = patronymicForName(firstName, male)
			}

			phone := normalizeRussianPhone(person.Phone)
			username := normalizeUsername(person.Username, firstName, lastName)

			p := Profile{
				FirstName:  firstName,
				LastName:   lastName,
				Patronymic: patronymic,
				Gender:     gender,
				Callsign:   callsigns[rand.Intn(len(callsigns))],
				Blood:      bloodTypes[rand.Intn(len(bloodTypes))],
				BirthDate:  birthDate,
				Age:        age,
				City:       cities[rand.Intn(len(cities))],
				Street:     streets[rand.Intn(len(streets))],
				House:      rand.Intn(178) + 1,
				Phone:      phone,
				Username:   username,
				Passport:   randomPassport(),
				AvatarURL:  avatarURL(style, male),
			}
			return p
		}
	}

	// Резервный локальный генератор: бот продолжает работать, даже если внешнее API недоступно.
	return generateLocalProfile(style)
}

func fetchFakePerson(gender string) (fakePerson, error) {
	q := url.Values{}
	q.Set("count", "1")
	q.Set("country", "ru")
	q.Set("gender", gender)
	q.Set("seed", strconv.FormatInt(time.Now().UnixNano(), 10))

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(fakeNamelyURL + "?" + q.Encode())
	if err != nil {
		return fakePerson{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fakePerson{}, fmt.Errorf("fakenamely HTTP %d", resp.StatusCode)
	}

	var payload fakeNamelyResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fakePerson{}, err
	}
	if !payload.Success || len(payload.Data) == 0 {
		return fakePerson{}, fmt.Errorf("fakenamely returned no data")
	}
	return payload.Data[0], nil
}

func fetchRealFaceURL(gender string) string {
	q := url.Values{}
	q.Set("gender", gender)
	q.Set("inc", "gender,picture")
	q.Set("noinfo", "1")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(randomUserURL + "?" + q.Encode())
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var payload randomUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || len(payload.Results) == 0 {
		return ""
	}
	return payload.Results[0].Picture.Large
}

func apiBirthDate(value string) (time.Time, int) {
	if value == "" {
		return time.Time{}, 0
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, 0
	}
	return t, calculateAge(t)
}

func calculateAge(birth time.Time) int {
	now := time.Now()
	age := now.Year() - birth.Year()
	if now.Month() < birth.Month() || (now.Month() == birth.Month() && now.Day() < birth.Day()) {
		age--
	}
	return age
}

func patronymicForName(first string, male bool) string {
	if value, ok := fatherPatronymics[first]; ok {
		if male {
			return value[0]
		}
		return value[1]
	}
	if male {
		return randomPatronymic()[0]
	}
	return randomPatronymic()[1]
}

func normalizeUsername(value, firstName, lastName string) string {
	value = strings.TrimPrefix(strings.TrimSpace(value), "@")
	value = nonUsername.ReplaceAllString(value, "")
	if len(value) >= 5 {
		return "@" + value
	}

	base := strings.ToLower(firstName + "_" + lastName)
	base = nonUsername.ReplaceAllString(base, "")
	return "@" + base + strconv.Itoa(rand.Intn(9000)+1000)
}

func normalizeRussianPhone(value string) string {
	digits := regexp.MustCompile(`\D+`).ReplaceAllString(value, "")
	if len(digits) >= 11 && strings.HasPrefix(digits, "7") {
		digits = digits[1:]
	} else if len(digits) >= 10 {
		digits = digits[len(digits)-10:]
	} else {
		digits = ""
	}

	if len(digits) != 10 || digits[0] != '9' {
		digits = "9" + randomDigits(9)
	}
	return fmt.Sprintf("+7 %s %s-%s-%s", digits[0:3], digits[3:6], digits[6:8], digits[8:10])
}

func randomDigits(n int) string {
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		b.WriteByte(byte('0' + rand.Intn(10)))
	}
	return b.String()
}

func generateLocalProfile(style string) Profile {
	isMale := rand.Intn(2) == 0
	var firstName, lastName, patronymic string
	var gender string
	pat := randomPatronymic()
	if isMale {
		gender = "Мужской"
		firstName = maleNames[rand.Intn(len(maleNames))]
		lastName = surnames[rand.Intn(len(surnames))]
		patronymic = pat[0]
	} else {
		gender = "Женский"
		firstName = femaleNames[rand.Intn(len(femaleNames))]
		lastName = feminineSurname(surnames[rand.Intn(len(surnames))])
		patronymic = pat[1]
	}

	birthDate, age := randomBirthDate()
	return Profile{
		FirstName:  firstName,
		LastName:   lastName,
		Patronymic: patronymic,
		Gender:     gender,
		Callsign:   callsigns[rand.Intn(len(callsigns))],
		Blood:      bloodTypes[rand.Intn(len(bloodTypes))],
		BirthDate:  birthDate,
		Age:        age,
		City:       cities[rand.Intn(len(cities))],
		Street:     streets[rand.Intn(len(streets))],
		House:      rand.Intn(178) + 1,
		Phone:      randomPhone(),
		Username:   randomUsername(),
		Passport:   randomPassport(),
		AvatarURL:  avatarURL(style, isMale),
	}
}

func feminineSurname(male string) string {
	// Формы на -ов/-ев/-ин/-ын/-ский/-цкий согласуются с женским полом.
	switch {
	case strings.HasSuffix(male, "ов"):
		return strings.TrimSuffix(male, "ов") + "ова"
	case strings.HasSuffix(male, "ев"):
		return strings.TrimSuffix(male, "ев") + "ева"
	case strings.HasSuffix(male, "ин"):
		return strings.TrimSuffix(male, "ин") + "ина"
	case strings.HasSuffix(male, "ын"):
		return strings.TrimSuffix(male, "ын") + "ына"
	case strings.HasSuffix(male, "ский"):
		return strings.TrimSuffix(male, "ский") + "ская"
	case strings.HasSuffix(male, "цкий"):
		return strings.TrimSuffix(male, "цкий") + "цкая"
	case strings.HasSuffix(male, "ой"):
		return strings.TrimSuffix(male, "ой") + "ая"
	default:
		// Гайдук, Пархоменко, Руденко, Кулиш и подобные фамилии не изменяются.
		return male
	}
}

func randomBirthDate() (time.Time, int) {
	now := time.Now()
	age := rand.Intn(31) + 20
	year := now.Year() - age
	month := time.Month(rand.Intn(12) + 1)
	day := rand.Intn(daysInMonth(year, month)) + 1
	birth := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	if birth.After(now) {
		birth = birth.AddDate(-1, 0, 0)
	}
	actualAge := now.Year() - birth.Year()
	if now.YearDay() < birth.YearDay() {
		actualAge--
	}
	if actualAge < 20 {
		birth = birth.AddDate(-1, 0, 0)
		actualAge++
	}
	if actualAge > 50 {
		birth = birth.AddDate(1, 0, 0)
		actualAge--
	}
	return birth, actualAge
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func randomPhone() string {
	return fmt.Sprintf("+7 9%02d %03d-%02d-%02d", rand.Intn(100), rand.Intn(1000), rand.Intn(100), rand.Intn(100))
}

func randomPassport() string {
	series := rand.Intn(9000) + 1000
	number := rand.Intn(900000) + 100000
	return fmt.Sprintf("%04d %06d", series, number)
}

func randomUsername() string {
	word := usernameWords[rand.Intn(len(usernameWords))]
	n := rand.Intn(9000) + 1000
	return "@" + word + strconv.Itoa(n)
}

func randomPatronymic() [2]string {
	keys := make([]string, 0, len(fatherPatronymics))
	for name := range fatherPatronymics {
		keys = append(keys, name)
	}
	return fatherPatronymics[keys[rand.Intn(len(keys))]]
}

func avatarURL(style string, male bool) string {
	seed := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
	switch style {
	case stylePixel:
		return fmt.Sprintf("https://api.dicebear.com/10.x/pixel-art/png?seed=%s&size=256", url.QueryEscape(seed))
	case styleCartoon:
		return fmt.Sprintf("https://api.dicebear.com/10.x/adventurer/png?seed=%s&size=256", url.QueryEscape(seed))
	case styleReal:
		if u := fetchRealFaceURL(map[bool]string{true: "male", false: "female"}[male]); u != "" {
			return u
		}
		gender := "women"
		if male {
			gender = "men"
		}
		return fmt.Sprintf("https://randomuser.me/api/portraits/%s/%d.jpg", gender, rand.Intn(100))
	default:
		return fmt.Sprintf("https://api.dicebear.com/10.x/adventurer/png?seed=%s&size=256", url.QueryEscape(seed))
	}
}

func formatProfile(p Profile) string {
	return fmt.Sprintf(
		"🪖 *ЛИЧНОЕ ДОСЬЕ*\n\n"+
			"👤 *ФИО:* %s\n"+
			"⚧ *Пол:* %s\n"+
			"🎖 *Позывной:* `%s`\n"+
			"🩸 *Группа крови:* %s\n"+
			"🎂 *Дата рождения:* %s\n"+
			"📅 *Возраст:* %d лет\n"+
			"📍 *Город:* %s\n"+
			"🏠 *Адрес:* %s, д\\. %d\n"+ // <-- Исправлено
			"📞 *Телефон:* `%s`\n"+
			"📱 *Username:* `%s`\n"+
			"🪪 *Паспорт тестовый:* `%s`\n\n"+
			"_Все данные в этом профиле синтетические и предназначены только для развлечения, тестов и демонстраций\\._",
		escapeMD(p.LastName+" "+p.FirstName+" "+p.Patronymic),
		escapeMD(p.Gender),
		escapeCode(p.Callsign),
		escapeMD(p.Blood),
		escapeMD(p.BirthDate.Format("02.01.2006")),
		p.Age,
		escapeMD(p.City),
		escapeMD(p.Street),
		p.House,
		escapeCode(p.Phone),
		escapeCode(p.Username),
		escapeCode(p.Passport),
	)
}

func escapeMD(s string) string {
	for _, ch := range []string{"\\", "_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"} {
		s = strings.ReplaceAll(s, ch, "\\"+ch)
	}
	return s
}

func escapeCode(s string) string {
	return strings.NewReplacer("\\", "\\\\", "`", "\\`").Replace(s)
}
