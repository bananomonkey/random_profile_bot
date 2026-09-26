package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	_ "modernc.org/sqlite"
)

const (
	styleReal     = "real"
	stylePixel    = "pixel"
	styleCartoon  = "cartoon"
	styleRobots   = "robots"
	styleInitials = "initials"
	styleBoring   = "boring"

	themeCyber    = "cyber"
	themeMilitary = "military"
	themeCorp     = "corp"

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
	Email      string
	Password   string
	Passport   string
	SNILS      string
	INN        string
	CardNumber string
	CardExpiry string
	CardCVV    string
	CardType   string
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

type BankCard struct {
	Number string
	Expiry string
	CVV    string
	Type   string
}

var nonUsername = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

var cyrToLat = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch",
	'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

var (
	cities = []string{
		"Самосир", "Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань",
		"Нижний Новгород", "Челябинск", "Самара", "Омск", "Ростов-на-Дону", "Уфа",
		"Красноярск", "Пермь", "Воронеж", "Волгоград", "Краснодар", "Саратов",
		"Тюмень", "Тольятти", "Ижевск", "Барнаул", "Ульяновск", "Иркутск",
		"Хабаровск", "Ярославль", "Владивосток", "Махачкала", "Томск", "Оренбург",
		"Кемерово", "Новокузнецк", "Рязань", "Астрахань", "Пенза", "Липецк",
		"Киров", "Чебоксары", "Тула", "Калининград", "Курск", "Сочи",
		"Ставрополь", "Белгород", "Брянск", "Владимир", "Архангельск", "Сургут",
		"Якутск", "Мурманск", "Грозный", "Бобруйск", "Урюпинск", "Жмеринка",
		"Выборг", "Таганрог", "Муром", "Псков", "Великий Новгород", "Смоленск", "Бахмут", "Симферополь", "Малая токмачка", "Чернобыль",
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
		"Фронтовая", "Героев", "Сиреневая", "Каштановая", "Луговая", "Озерная",
		"Горная", "Высокая", "Верхняя", "Нижняя", "Каменская", "Станционная",
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
		"Яйцеслав", "Акакий", "Ануфрий", "Пантелеймон", "Архип", "Кузьма", "Лукьян",
		"Поликарп", "Никанор", "Евстигней", "Епифан", "Кондрат", "Герасим", "Спиридон",
		"Порфирий", "Самсон", "Евдоким", "Нестор", "Парамон", "Савва", "Митрофан",
		"Гаврила", "Ермолай", "Харлампий", "Михей", "Севастьян", "Илларион", "Галактион",
		"Венедикт", "Сидор", "Аверьян", "Афанасий", "Лаврентий", "Меркурий", "Пафнутий",
		"Серафим", "Фома", "Софрон", "Терентий", "Федот", "Гурий", "Климент", "Макарий",
		"Мефодий", "Вакула", "Корней", "Ермак", "Игнатий", "Наум", "Орест", "Прокофий",
		"Радион", "Савватий", "Трифон", "Феофан", "Филимон", "Харитон", "Яков", "Яромир",
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
		"Майя", "Мелания", "Нина", "Прасковья", "Рада", "Серафима", "Стефания",
		"Элина", "Эмилия", "Эльвира", "Аксинья", "Алевтина", "Белла", "Виолетта", "Дина",
		"Зоя", "Лиана", "Луиза", "Марта", "Розалия", "Сусанна", "Эвелина", "Элеонора",
		"Акулина", "Пелагея", "Евдокия", "Фекла", "Устинья", "Матрена", "Аграфена",
		"Валентина", "Антонина", "Таиса", "Серафима", "Марфа", "Олимпиада", "Ираида",
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
		"Туманов", "Фокин", "Шувалов", "Гусев", "Кузьмин", "Кулиш", "Мазур", "Савин", "Сафронов",
		"Лыткин", "Мороз", "Скориков", "Белкин", "Серов", "Безруков", "Боярский", "Дроздов",
		"Дубровский", "Ермаков", "Золотарев", "Казанцев", "Коновалов", "Корнилов", "Лазарев",
		"Муратов", "Немцов", "Орехов", "Пешков", "Разумовский", "Соловьев", "Тургенев",
	}

	callsigns = []string{
		"Фембой", "Нефор", "Кукич", "Дилдак", "Пельмень", "Чебурек", "Огурчик", "Банан",
		"Кисель", "Тапок", "Хряк", "Шмыга", "Шуруп", "Булка", "Карапуз", "Жирчик", "Пончик",
		"Колбаса", "Чушпан", "Суета", "Дрель", "Джигурда", "Шнырь", "Шрек", "Груздь",
		"Сосиска", "Пельмешка", "Борщ", "Тюлень", "Колобок", "Котлета", "Фрикаделька", "Точик",
		"Альтушка", "Скуф", "Тюбик", "Масик", "Чебурашка", "Копатыч", "Чепух", "Шеф",
		"Гигачад", "Потужный", "Доширак", "Шаурма", "Каблук", "Хомяк", "Телепузик",
		"Жид", "Шлепа", "Босс", "Чиназес", "База", "Пайп", "Волына", "Батон",
		"Сухарик", "Ультра-аморал", "Бублик", "Сырник", "Коржик", "Компот", "Степашка", "Лаваш",
		"Лосик", "Блинчик", "Беляш", "Самса", "Хинкаль", "Вареник", "Шашлык", "Драник",
		"Кефир", "Сметанец", "Майонез", "Негр", "Душнила", "Инцел", "Всевышний", "Смешарик",
		"Гой", "Гойда", "Движуха", "СВОйный", "Соя",
		"Гранит", "Седой", "Ворон", "Чекист", "Алтай", "Скиф", "Шаман", "Туман", "Варяг",
		"Компас", "Сструна", "Лютый", "Кремень", "Гром", "Борз", "Маэстро", "Барс", "Тайфун",
		"Шторм", "Буран", "Ветер", "Север", "Юг", "Восток", "Запад", "Феникс", "Сокол",
		"Ястреб", "Коршун", "Орёл", "Рысь", "Тигр", "Волк", "Медведь", "Кедр", "Тополь",
		"Сапсан", "Кобра", "Гюрза", "Удав", "Шершень", "Шмель", "Беркут", "Кондор", "Клык",
		"Коготь", "Бритва", "Молот", "Кувалда", "Нож", "Искра", "Пламя", "Уголь", "Дым",
		"Залп", "Калибр", "Вектор", "Радар", "Сигнал", "Маяк", "Форт", "Редут", "Щит",
		"Дозор", "Скала", "Утёс", "Пик", "Вихрь", "Циклон", "Гроза", "Молния", "Смерч",
		"Зенит", "Орион", "Тайга", "Полюс", "Байкал", "Енисей", "Дон", "Волга", "Амур",
		"Талисман", "Карат", "Рубеж", "Призрак", "Тень", "Мираж", "Фантом", "Титан", "Атлас", "Прометей",
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
		"femboy", "nefor", "kukich", "dildak", "skuf", "gigachad", "abobus", "shlepa", "chinazes",
	}

	fatherPatronymics = map[string][2]string{
		"Александр": {"Александрович", "Александровна"}, "Алексей": {"Алексеевич", "Алексеевна"},
		"Андрей": {"Андреевич", "Андреевна"}, "Антон": {"Антонович", "Антоновна"},
		"Аркадий": {"Аркадьевич", "Аркадьевна"}, "Артём": {"Артэмович", "Артёмовна"},
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
		"Яйцеслав": {"Яйцеславович", "Яйцеславовна"}, "Акакий": {"Акакиевич", "Акакиевна"},
		"Ануфрий": {"Ануфриевич", "Ануфриевна"}, "Пантелеймон": {"Пантелеймонович", "Пантелеймоновна"},
		"Архип": {"Архипович", "Архиповна"}, "Кузьма": {"Кузьмич", "Кузьминична"},
		"Лукьян": {"Лукьянович", "Лукьяновна"}, "Поликарп": {"Поликарпович", "Поликарповна"},
		"Никанор": {"Никанорович", "Никаноровна"}, "Евстигней": {"Евстигнеевич", "Евстигнеевна"},
		"Епифан": {"Епифанович", "Епифановна"}, "Кондрат": {"Кондратович", "Кондратовна"},
		"Герасим": {"Герасимович", "Герасимовна"}, "Спиридон": {"Спиридонович", "Спиридоновна"},
		"Порфирий": {"Порфирьевич", "Порфирьевна"}, "Самсон": {"Самсонович", "Самсоновна"},
		"Евдоким": {"Евдокимович", "Евдокимовна"}, "Нестор": {"Несторович", "Несторовна"},
		"Парамон": {"Парамонович", "Парамоновна"}, "Савва": {"Саввич", "Саввична"},
		"Митрофан": {"Митрофанович", "Митрофановна"}, "Гаврила": {"Гаврилович", "Гавриловна"},
		"Ермолай": {"Ермолаевич", "Ермолаевна"}, "Харлампий": {"Харлампиевич", "Харлампиевна"},
		"Михей": {"Михеевич", "Михеевна"}, "Севастьян": {"Севастьянович", "Севастьяновна"},
		"Илларион": {"Илларионович", "Илларионовна"}, "Галактион": {"Галактионович", "Галактионовна"},
		"Венедикт": {"Венедиктович", "Венедиктовна"}, "Сидор": {"Сидорович", "Сидоровна"},
		"Аверьян": {"Аверьянович", "Аверьяновна"}, "Афанасий": {"Афанасьевич", "Афанасьевна"},
		"Лаврентий": {"Лаврентьевич", "Лаврентьевна"}, "Меркурий": {"Меркурьевич", "Меркурьевна"},
		"Пафнутий": {"Пафнутьевич", "Пафнутьевна"}, "Серафим": {"Серафимович", "Серафимовна"},
		"Фома": {"Фомич", "Фоминична"}, "Софрон": {"Софронович", "Софроновна"},
		"Терентий": {"Терентьевич", "Терентьевна"}, "Федот": {"Федотович", "Федотовна"},
		"Гурий": {"Гурьевич", "Гурьевна"}, "Климент": {"Климентович", "Климентовна"},
		"Макарий": {"Макарьевич", "Макарьевна"}, "Мефодий": {"Мефодьевич", "Мефодьевна"},
		"Вакула": {"Вакулович", "Вакуловна"}, "Корней": {"Корнеевич", "Корнеевна"},
		"Ермак": {"Ермакович", "Ермаковна"}, "Игнатий": {"Игнатьевич", "Игнатьевна"},
		"Наум": {"Наумович", "Наумовна"}, "Орест": {"Орестович", "Орестовна"},
		"Прокофий": {"Прокофьевич", "Прокофьевна"}, "Радион": {"Радионович", "Радионовна"},
		"Савватий": {"Савватьевич", "Савватьевна"}, "Трифон": {"Трифнович", "Трифновна"},
		"Феофан": {"Феофанович", "Феофановна"}, "Филимон": {"Филимонович", "Филимоновна"},
		"Харитон": {"Харитонович", "Харитоновна"}, "Яков": {"Яковлевич", "Яковлевна"},
		"Яромир": {"Яромирович", "Яромировна"},
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

	// Фоновое удаление устаревших профилей из памяти
	startProfileCleaner()

	// Регистрируем меню команд в левом нижнем углу Telegram
	cmdConfig := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "start", Description: "🚀 Главное меню и запуск"},
		tgbotapi.BotCommand{Command: "generate", Description: "🎲 Сгенерировать профиль"},
		tgbotapi.BotCommand{Command: "favorites", Description: "⭐ Избранные персонажи"},
		tgbotapi.BotCommand{Command: "settings", Description: "⚙️ Настройки стиля и темы"},
		tgbotapi.BotCommand{Command: "help", Description: "ℹ️ Справка по боту"},
	)
	if _, err := bot.Request(cmdConfig); err != nil {
		log.Printf("ошибка установки команд бота: %v", err)
	}

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
			avatar_style TEXT NOT NULL DEFAULT 'pixel',
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		);
	`)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	// Миграция: колонка с темой пропуска (игнорируем ошибку, если уже существует).
	_, _ = db.Exec("ALTER TABLE user_settings ADD COLUMN pass_theme TEXT NOT NULL DEFAULT 'cyber'")

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS favorites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			passport TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
			UNIQUE(user_id, passport)
		);
	`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create favorites table: %w", err)
	}

	return db, nil
}

func handleMessage(bot *tgbotapi.BotAPI, db *sql.DB, msg *tgbotapi.Message) {
	if msg.Text == "" {
		return
	}

	chatID := msg.Chat.ID
	userID := msg.From.ID

	switch msg.Text {
	case "/start":
		welcomeText := "🪖 *Добро пожаловать в Генератор Профилей\\!*\n\n" +
			"Этот бот создаёт полностью вымышленные личности для игр, ролевых проектов и тестирования:\n" +
			"• 👤 *ФИО, Позывной и Паспортные данные*\n" +
			"• 📍 *Город и Группа крови*\n" +
			"• 🖼 *Уникальный аватар* \\(пиксельный, реалистичный, роботы и др\\.\\)\n" +
			"• 🎫 *Отрисовка карточки пропуска* по кнопке под анкетой\n\n" +
			"Нажми *Сгенерировать профиль*, чтобы начать\\!"

		sendMsg := tgbotapi.NewMessage(chatID, welcomeText)
		sendMsg.ParseMode = tgbotapi.ModeMarkdownV2
		sendMsg.ReplyMarkup = mainKeyboard()
		_, _ = bot.Send(sendMsg)

	case "/generate", "🎲 Сгенерировать профиль":
		generateAndSendProfile(bot, db, chatID, userID)

	case "/settings", "⚙️ Настройки":
		style, err := getUserStyle(db, userID)
		if err != nil {
			log.Printf("get style: %v", err)
			return
		}
		theme, _ := getUserPassTheme(db, userID)
		sendMsg := tgbotapi.NewMessage(chatID, settingsText(style, theme))
		sendMsg.ParseMode = tgbotapi.ModeMarkdownV2
		sendMsg.ReplyMarkup = settingsKeyboard(style, theme)
		_, _ = bot.Send(sendMsg)

	case "/favorites", "⭐ Избранное":
		sendFavorites(bot, db, chatID, userID)

	case "/help":
		helpText := "ℹ️ *Справка*\n\n" +
			"• `/generate` — сгенерировать новую карточку персонажа\n" +
			"• `/settings` — выбрать стиль аватара и тему пропуска\n" +
			"• `/favorites` — список сохранённых персонажей\n" +
			"• Под каждой карточкой есть кнопка *🎫 Сгенерировать пропуск*, которая создаст график\\-карточку именно для этого персонажа\\."

		sendMsg := tgbotapi.NewMessage(chatID, helpText)
		sendMsg.ParseMode = tgbotapi.ModeMarkdownV2
		sendMsg.ReplyMarkup = mainKeyboard()
		_, _ = bot.Send(sendMsg)

	default:
		sendMsg := tgbotapi.NewMessage(chatID, "Используйте кнопки меню или команду /generate.")
		sendMsg.ReplyMarkup = mainKeyboard()
		_, _ = bot.Send(sendMsg)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, db *sql.DB, callback *tgbotapi.CallbackQuery) {
	answer := tgbotapi.NewCallback(callback.ID, "")
	if _, err := bot.Request(answer); err != nil {
		log.Printf("answer callback: %v", err)
	}

	chatID := callback.Message.Chat.ID
	userID := callback.From.ID

	switch {
	// Пропуск для конкретного паспорта
	case strings.HasPrefix(callback.Data, "pass_"):
		passport := strings.TrimPrefix(callback.Data, "pass_")
		profile, ok := getStoredProfile(passport)
		if !ok {
			alert := tgbotapi.NewCallbackWithAlert(callback.ID, "Профиль устарел или не найден! Сгенерируйте новый.")
			_, _ = bot.Request(alert)
			return
		}
		sendPassImage(bot, db, chatID, userID, profile)
		return

	// Добавление профиля в избранное
	case strings.HasPrefix(callback.Data, "fav_"):
		passport := strings.TrimPrefix(callback.Data, "fav_")
		profile, ok := getStoredProfile(passport)
		if !ok {
			alert := tgbotapi.NewCallbackWithAlert(callback.ID, "Профиль устарел или не найден! Сгенерируйте новый.")
			_, _ = bot.Request(alert)
			return
		}
		text := "⭐ Добавлено в избранное!"
		if err := saveFavorite(db, userID, profile); err != nil {
			log.Printf("save favorite: %v", err)
			text = "Не удалось сохранить в избранное"
		}
		alert := tgbotapi.NewCallbackWithAlert(callback.ID, text)
		_, _ = bot.Request(alert)
		return

	// Пропуск из избранного по id записи
	case strings.HasPrefix(callback.Data, "favpass_"):
		id, err := strconv.ParseInt(strings.TrimPrefix(callback.Data, "favpass_"), 10, 64)
		if err != nil {
			return
		}
		profile, ok := getFavorite(db, userID, id)
		if !ok {
			alert := tgbotapi.NewCallbackWithAlert(callback.ID, "Персонаж не найден в избранном.")
			_, _ = bot.Request(alert)
			return
		}
		sendPassImage(bot, db, chatID, userID, profile)
		return
	}

	switch callback.Data {
	case "generate_profile":
		generateAndSendProfile(bot, db, chatID, userID)
	case "favorites":
		sendFavorites(bot, db, chatID, userID)
	case "settings":
		style, err := getUserStyle(db, userID)
		if err != nil {
			log.Printf("get style: %v", err)
			return
		}
		theme, _ := getUserPassTheme(db, userID)
		sendOrEditMessage(bot, chatID, callback.Message, settingsText(style, theme), settingsKeyboard(style, theme))
	case "style_real", "style_pixel", "style_cartoon", "style_robots", "style_initials", "style_boring":
		style := strings.TrimPrefix(callback.Data, "style_")
		if err := setUserStyle(db, userID, style); err != nil {
			log.Printf("set style: %v", err)
			return
		}
		theme, _ := getUserPassTheme(db, userID)
		sendOrEditMessage(bot, chatID, callback.Message, settingsText(style, theme), settingsKeyboard(style, theme))
	case "theme_cyber", "theme_military", "theme_corp":
		theme := strings.TrimPrefix(callback.Data, "theme_")
		if err := setUserPassTheme(db, userID, theme); err != nil {
			log.Printf("set theme: %v", err)
			return
		}
		style, _ := getUserStyle(db, userID)
		sendOrEditMessage(bot, chatID, callback.Message, settingsText(style, theme), settingsKeyboard(style, theme))
	case "back":
		text := "🪖 *Генератор профилей*\n\nНажми кнопку ниже — бот создаст новый полностью вымышленный профиль\\."
		sendOrEditMessage(bot, chatID, callback.Message, text, mainKeyboard())
	}
}

func sendPassImage(bot *tgbotapi.BotAPI, db *sql.DB, chatID, userID int64, profile Profile) {
	theme, _ := getUserPassTheme(db, userID)

	passBytes, err := GeneratePassImage(profile, theme)
	if err != nil {
		log.Printf("ошибка генерации пропуска: %v", err)
		return
	}

	fileBytes := tgbotapi.FileBytes{
		Name:  "pass.png",
		Bytes: passBytes,
	}
	photo := tgbotapi.NewPhoto(chatID, fileBytes)
	photo.Caption = fmt.Sprintf("🎫 *Пропуск для %s %s*", escapeMD(profile.FirstName), escapeMD(profile.LastName))
	photo.ParseMode = tgbotapi.ModeMarkdownV2
	if _, err := bot.Send(photo); err != nil {
		log.Printf("send pass image error: %v", err)
	}
}

func sendFavorites(bot *tgbotapi.BotAPI, db *sql.DB, chatID, userID int64) {
	favorites, err := getFavorites(db, userID)
	if err != nil {
		log.Printf("get favorites: %v", err)
		return
	}

	if len(favorites) == 0 {
		msg := tgbotapi.NewMessage(chatID, "⭐ *Избранное пусто\\.*\n\nСгенерируйте профиль и нажмите «⭐ В избранное» под карточкой\\.")
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = mainKeyboard()
		_, _ = bot.Send(msg)
		return
	}

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(favorites)+1)
	for _, f := range favorites {
		label := fmt.Sprintf("🎫 %s %s (%s)", f.Profile.LastName, f.Profile.FirstName, f.Profile.Passport)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "favpass_"+strconv.FormatInt(f.ID, 10)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back"),
	))

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⭐ *Избранные персонажи* \\(%d\\)\n\nНажмите на персонажа, чтобы получить его пропуск\\.", len(favorites)))
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	_, _ = bot.Send(msg)
}

func sendOrEditMessage(bot *tgbotapi.BotAPI, chatID int64, message *tgbotapi.Message, text string, markup tgbotapi.InlineKeyboardMarkup) {
	if message != nil && len(message.Photo) > 0 {
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, message.MessageID)
		_, _ = bot.Request(deleteMsg)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = markup
		if _, err := bot.Send(msg); err != nil {
			log.Printf("send new message error: %v", err)
		}
		return
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, message.MessageID, text, markup)
	edit.ParseMode = tgbotapi.ModeMarkdownV2
	if _, err := bot.Send(edit); err != nil {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = markup
		_, _ = bot.Send(msg)
	}
}

func generateAndSendProfile(bot *tgbotapi.BotAPI, db *sql.DB, chatID, userID int64) {
	style, err := getUserStyle(db, userID)
	if err != nil {
		log.Printf("get style before generation: %v", err)
		style = stylePixel
	}

	profile := generateProfile(style)

	// Сохраняем профиль по уникальному номеру паспорта (с TTL)
	storeProfile(profile)

	caption := formatProfile(profile)
	keyboard := profileKeyboard(profile.Passport)

	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(profile.AvatarURL))
	photo.Caption = caption
	photo.ParseMode = tgbotapi.ModeMarkdownV2
	photo.ReplyMarkup = keyboard

	if _, err := bot.Send(photo); err != nil {
		log.Printf("send profile photo: %v", err)
		msg := tgbotapi.NewMessage(chatID, caption)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		msg.ReplyMarkup = keyboard
		if _, sendErr := bot.Send(msg); sendErr != nil {
			log.Printf("send profile text fallback: %v", sendErr)
		}
	}
}

// Хранилище сгенерированных профилей по их номеру паспорта.
// Записи живут ограниченное время (profileTTL), чтобы не забивать память.
type storedProfile struct {
	Profile   Profile
	CreatedAt time.Time
}

const profileTTL = 24 * time.Hour

var (
	profilesMutex  sync.RWMutex
	storedProfiles = make(map[string]storedProfile)
)

func storeProfile(p Profile) {
	profilesMutex.Lock()
	storedProfiles[p.Passport] = storedProfile{Profile: p, CreatedAt: time.Now()}
	profilesMutex.Unlock()
}

func getStoredProfile(passport string) (Profile, bool) {
	profilesMutex.RLock()
	sp, ok := storedProfiles[passport]
	profilesMutex.RUnlock()
	if !ok || time.Since(sp.CreatedAt) > profileTTL {
		return Profile{}, false
	}
	return sp.Profile, true
}

// startProfileCleaner раз в час удаляет устаревшие профили из памяти.
func startProfileCleaner() {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().Add(-profileTTL)
			profilesMutex.Lock()
			for passport, sp := range storedProfiles {
				if sp.CreatedAt.Before(cutoff) {
					delete(storedProfiles, passport)
				}
			}
			profilesMutex.Unlock()
		}
	}()
}

// Клавиатура привязывает кнопку пропуска к конкретному паспорту
func profileKeyboard(passport string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎫 Сгенерировать пропуск", "pass_"+passport),
			tgbotapi.NewInlineKeyboardButtonData("⭐ В избранное", "fav_"+passport),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Сгенерировать ещё", "generate_profile"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Настройки", "settings"),
		),
	)
}

func mainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🪪 Сгенерировать профиль", "generate_profile"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ Избранное", "favorites"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Настройки", "settings"),
		),
	)
}

func settingsKeyboard(style, theme string) tgbotapi.InlineKeyboardMarkup {
	button := func(text, data, current, value string) tgbotapi.InlineKeyboardButton {
		if current == value {
			text = "✅ " + text
		}
		return tgbotapi.NewInlineKeyboardButtonData(text, data)
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			button("Реальное лицо", "style_real", style, styleReal),
			button("Пиксельный", "style_pixel", style, stylePixel),
		),
		tgbotapi.NewInlineKeyboardRow(
			button("Мультяшный", "style_cartoon", style, styleCartoon),
			button("Роботы", "style_robots", style, styleRobots),
		),
		tgbotapi.NewInlineKeyboardRow(
			button("Инициалы", "style_initials", style, styleInitials),
			button("Абстрактный", "style_boring", style, styleBoring),
		),
		tgbotapi.NewInlineKeyboardRow(
			button("Cyberpunk", "theme_cyber", theme, themeCyber),
			button("Military", "theme_military", theme, themeMilitary),
			button("Corporate", "theme_corp", theme, themeCorp),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back"),
		),
	)
}

func settingsText(style, theme string) string {
	current := map[string]string{
		styleReal:     "📷 Реальное лицо",
		stylePixel:    "👾 Пиксельный аватар",
		styleCartoon:  "🎨 Мультяшный / Векторный",
		styleRobots:   "🤖 Роботы (Robohash)",
		styleInitials: "🔤 Инициалы (UI Avatars)",
		styleBoring:   "🎨 Абстракция (Boring Avatars)",
	}[style]
	if current == "" {
		current = "👾 Пиксельный аватар"
	}

	currentTheme := map[string]string{
		themeCyber:    "🌃 Cyberpunk / Neon",
		themeMilitary: "🪖 Military / Tactical",
		themeCorp:     "🏢 Corporate / Modern",
	}[theme]
	if currentTheme == "" {
		currentTheme = "🌃 Cyberpunk / Neon"
	}

	return fmt.Sprintf(
		"⚙️ *Настройки*\n\nТекущий стиль аватара: *%s*\nТема пропуска: *%s*\n\nВыбери параметры для следующей генерации\\.",
		escapeMD(current), escapeMD(currentTheme),
	)
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
	valid := map[string]bool{
		styleReal: true, stylePixel: true, styleCartoon: true,
		styleRobots: true, styleInitials: true, styleBoring: true,
	}
	if !valid[style] {
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

func getUserPassTheme(db *sql.DB, userID int64) (string, error) {
	var theme string
	err := db.QueryRow("SELECT pass_theme FROM user_settings WHERE user_id = ?", userID).Scan(&theme)
	if err == sql.ErrNoRows {
		return themeCyber, nil
	}
	if err != nil || theme == "" {
		return themeCyber, err
	}
	return theme, nil
}

func setUserPassTheme(db *sql.DB, userID int64, theme string) error {
	valid := map[string]bool{themeCyber: true, themeMilitary: true, themeCorp: true}
	if !valid[theme] {
		theme = themeCyber
	}
	_, err := db.Exec(`
		INSERT INTO user_settings (user_id, avatar_style, pass_theme, updated_at)
		VALUES (?, ?, ?, strftime('%s','now'))
		ON CONFLICT(user_id) DO UPDATE SET
			pass_theme = excluded.pass_theme,
			updated_at = strftime('%s','now')
	`, userID, stylePixel, theme)
	return err
}

type favorite struct {
	ID      int64
	Profile Profile
}

func saveFavorite(db *sql.DB, userID int64, p Profile) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO favorites (user_id, passport, data, created_at)
		VALUES (?, ?, ?, strftime('%s','now'))
		ON CONFLICT(user_id, passport) DO UPDATE SET
			data = excluded.data,
			created_at = strftime('%s','now')
	`, userID, p.Passport, string(data))
	return err
}

func getFavorites(db *sql.DB, userID int64) ([]favorite, error) {
	rows, err := db.Query("SELECT id, data FROM favorites WHERE user_id = ? ORDER BY created_at DESC, id DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []favorite
	for rows.Next() {
		var (
			id   int64
			data string
		)
		if err := rows.Scan(&id, &data); err != nil {
			return nil, err
		}
		var p Profile
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			continue
		}
		favorites = append(favorites, favorite{ID: id, Profile: p})
	}
	return favorites, rows.Err()
}

func getFavorite(db *sql.DB, userID, id int64) (Profile, bool) {
	var data string
	err := db.QueryRow("SELECT data FROM favorites WHERE id = ? AND user_id = ?", id, userID).Scan(&data)
	if err != nil {
		return Profile{}, false
	}
	var p Profile
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return Profile{}, false
	}
	return p, true
}

func generateProfile(style string) Profile {
	genderAPI := "male"
	if rand.Intn(2) == 1 {
		genderAPI = "female"
	}

	card := generateBankCard()

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
			username := generateTGUsername(firstName, lastName)
			email := randomEmail(username)
			password := randomPassword()

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
				Email:      email,
				Password:   password,
				Passport:   randomPassport(),
				SNILS:      generateSNILS(),
				INN:        generateINN(),
				CardNumber: card.Number,
				CardExpiry: card.Expiry,
				CardCVV:    card.CVV,
				CardType:   card.Type,
				AvatarURL:  avatarURL(style, male, firstName, lastName),
			}
			return p
		}
	}

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

func transliterate(s string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(s) {
		if lat, ok := cyrToLat[r]; ok {
			sb.WriteString(lat)
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func generateTGUsername(firstName, lastName string) string {
	fn := transliterate(firstName)
	ln := transliterate(lastName)

	patterns := []string{}

	if len(fn) >= 3 && len(ln) >= 3 {
		patterns = append(patterns, fn+"_"+ln)
		patterns = append(patterns, ln+"_"+fn)
		patterns = append(patterns, fn+"_"+string(ln[0]))
		patterns = append(patterns, string(fn[0])+"_"+ln)
	}

	if len(fn) >= 3 {
		patterns = append(patterns, fn+strconv.Itoa(rand.Intn(900)+100))
		patterns = append(patterns, fn+"_"+strconv.Itoa(rand.Intn(90)+10))
	}

	if len(ln) >= 3 {
		patterns = append(patterns, ln+strconv.Itoa(rand.Intn(900)+100))
		patterns = append(patterns, ln+"_"+strconv.Itoa(rand.Intn(90)+10))
	}

	word := usernameWords[rand.Intn(len(usernameWords))]
	patterns = append(patterns, word+"_"+strconv.Itoa(rand.Intn(900)+100))

	username := patterns[rand.Intn(len(patterns))]
	username = nonUsername.ReplaceAllString(username, "")

	for len(username) < 5 {
		username += strconv.Itoa(rand.Intn(10))
	}

	return "@" + strings.ToLower(username)
}

func randomEmail(username string) string {
	cleanUser := strings.TrimPrefix(username, "@")
	domains := []string{"gmail.com", "yandex.ru", "mail.ru", "rambler.ru", "bk.ru", "inbox.ru", "icloud.com"}
	domain := domains[rand.Intn(len(domains))]
	return cleanUser + "@" + domain
}

func randomPassword() string {
	const (
		lowerChars   = "abcdefghijklmnopqrstuvwxyz"
		upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digitChars   = "0123456789"
		specialChars = "!@#$%^&*"
		allChars     = lowerChars + upperChars + digitChars + specialChars
	)
	length := rand.Intn(5) + 10
	b := make([]byte, length)
	b[0] = lowerChars[rand.Intn(len(lowerChars))]
	b[1] = upperChars[rand.Intn(len(upperChars))]
	b[2] = digitChars[rand.Intn(len(digitChars))]
	b[3] = specialChars[rand.Intn(len(specialChars))]

	for i := 4; i < length; i++ {
		b[i] = allChars[rand.Intn(len(allChars))]
	}
	rand.Shuffle(len(b), func(i, j int) { b[i], b[j] = b[j], b[i] })
	return string(b)
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

// Генератор валидного СНИЛС по контрольной сумме
func generateSNILS() string {
	digits := make([]int, 9)
	for i := 0; i < 9; i++ {
		digits[i] = rand.Intn(10)
	}

	sum := 0
	for i := 0; i < 9; i++ {
		sum += digits[i] * (9 - i)
	}

	var checkSum int
	if sum < 100 {
		checkSum = sum
	} else if sum == 100 || sum == 101 {
		checkSum = 0
	} else {
		rem := sum % 101
		if rem == 100 || rem == 101 {
			checkSum = 0
		} else {
			checkSum = rem
		}
	}

	return fmt.Sprintf("%d%d%d-%d%d%d-%d%d%d %02d",
		digits[0], digits[1], digits[2],
		digits[3], digits[4], digits[5],
		digits[6], digits[7], digits[8],
		checkSum)
}

// Генератор валидного 12-значного ИНН физического лица
func generateINN() string {
	digits := make([]int, 12)
	regions := []int{77, 50, 78, 16, 61, 23, 66, 52, 36, 54}
	reg := regions[rand.Intn(len(regions))]
	digits[0] = reg / 10
	digits[1] = reg % 10

	for i := 2; i < 10; i++ {
		digits[i] = rand.Intn(10)
	}

	mult1 := []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum1 := 0
	for i := 0; i < 10; i++ {
		sum1 += digits[i] * mult1[i]
	}
	digits[10] = (sum1 % 11) % 10

	mult2 := []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	sum2 := 0
	for i := 0; i < 11; i++ {
		sum2 += digits[i] * mult2[i]
	}
	digits[11] = (sum2 % 11) % 10

	var sb strings.Builder
	for _, d := range digits {
		sb.WriteString(strconv.Itoa(d))
	}
	return sb.String()
}

// Генератор банковских карт по алгоритму Луна
func generateBankCard() BankCard {
	types := []string{"МИР", "Visa", "Mastercard"}
	cardType := types[rand.Intn(len(types))]

	var prefix []int
	switch cardType {
	case "МИР":
		prefix = []int{2, 2, 0, 0}
	case "Visa":
		prefix = []int{4}
	case "Mastercard":
		prefix = []int{5, 2}
	}

	digits := make([]int, 16)
	copy(digits, prefix)
	for i := len(prefix); i < 15; i++ {
		digits[i] = rand.Intn(10)
	}

	sum := 0
	for i := 0; i < 15; i++ {
		d := digits[i]
		if i%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	digits[15] = (10 - (sum % 10)) % 10

	var sb strings.Builder
	for i, d := range digits {
		if i > 0 && i%4 == 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(strconv.Itoa(d))
	}

	now := time.Now()
	expYear := (now.Year() % 100) + rand.Intn(4) + 1
	expMonth := rand.Intn(12) + 1
	expiry := fmt.Sprintf("%02d/%02d", expMonth, expYear)
	cvv := fmt.Sprintf("%03d", rand.Intn(1000))

	return BankCard{
		Number: sb.String(),
		Expiry: expiry,
		CVV:    cvv,
		Type:   cardType,
	}
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
	username := generateTGUsername(firstName, lastName)
	email := randomEmail(username)
	password := randomPassword()
	card := generateBankCard()

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
		Username:   username,
		Email:      email,
		Password:   password,
		Passport:   randomPassport(),
		SNILS:      generateSNILS(),
		INN:        generateINN(),
		CardNumber: card.Number,
		CardExpiry: card.Expiry,
		CardCVV:    card.CVV,
		CardType:   card.Type,
		AvatarURL:  avatarURL(style, isMale, firstName, lastName),
	}
}

func feminineSurname(male string) string {
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

func randomPatronymic() [2]string {
	keys := make([]string, 0, len(fatherPatronymics))
	for name := range fatherPatronymics {
		keys = append(keys, name)
	}
	return fatherPatronymics[keys[rand.Intn(len(keys))]]
}

func avatarURL(style string, male bool, firstName, lastName string) string {
	seed := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
	switch style {
	case stylePixel:
		return fmt.Sprintf("https://api.dicebear.com/10.x/pixel-art/png?seed=%s&size=256", url.QueryEscape(seed))
	case styleCartoon:
		return fmt.Sprintf("https://api.dicebear.com/10.x/adventurer/png?seed=%s&size=256", url.QueryEscape(seed))
	case styleRobots:
		return fmt.Sprintf("https://robohash.org/%s.png?set=set1&size=256x256", url.QueryEscape(seed))
	case styleInitials:
		fullName := url.QueryEscape(fmt.Sprintf("%s %s", firstName, lastName))
		return fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=random&color=fff&size=256", fullName)
	case styleBoring:
		return fmt.Sprintf("https://source.boringavatars.com/beam/256/%s", url.QueryEscape(seed))
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
			"🏠 *Адрес:* %s, д\\. %d\n"+
			"📞 *Телефон:* `%s`\n"+
			"📱 *Username:* `%s`\n"+
			"📧 *Email:* `%s`\n"+
			"🔑 *Пароль:* `%s`\n"+
			"🪪 *Паспорт:* `%s`\n"+
			"📜 *СНИЛС:* `%s`\n"+
			"📑 *ИНН:* `%s`\n"+
			"💳 *Карта \\(%s\\):* `%s`\n"+
			"⏳ *Срок:* `%s` \\| *CVV:* `%s`\n\n"+
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
		escapeCode(p.Email),
		escapeCode(p.Password),
		escapeCode(p.Passport),
		escapeCode(p.SNILS),
		escapeCode(p.INN),
		escapeMD(p.CardType),
		escapeCode(p.CardNumber),
		escapeCode(p.CardExpiry),
		escapeCode(p.CardCVV),
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

// --- Генерация изображения пропуска ---

type passTheme struct {
	bg     color.RGBA
	header color.RGBA
	border color.RGBA
	title  color.RGBA
	text   color.RGBA
	muted  color.RGBA
	accent color.RGBA
	danger color.RGBA
	qrFg   color.RGBA
	qrBg   color.RGBA
}

func themeColors(theme string) passTheme {
	switch theme {
	case themeMilitary:
		return passTheme{
			bg:     color.RGBA{42, 48, 34, 255},
			header: color.RGBA{74, 84, 54, 255},
			border: color.RGBA{96, 106, 72, 255},
			title:  color.RGBA{226, 232, 208, 255},
			text:   color.RGBA{222, 226, 206, 255},
			muted:  color.RGBA{176, 182, 150, 255},
			accent: color.RGBA{196, 214, 120, 255},
			danger: color.RGBA{224, 120, 90, 255},
			qrFg:   color.RGBA{15, 18, 10, 255},
			qrBg:   color.RGBA{255, 255, 255, 255},
		}
	case themeCorp:
		return passTheme{
			bg:     color.RGBA{245, 247, 250, 255},
			header: color.RGBA{33, 64, 120, 255},
			border: color.RGBA{200, 208, 218, 255},
			title:  color.RGBA{255, 255, 255, 255},
			text:   color.RGBA{30, 38, 50, 255},
			muted:  color.RGBA{110, 120, 135, 255},
			accent: color.RGBA{33, 110, 200, 255},
			danger: color.RGBA{190, 60, 50, 255},
			qrFg:   color.RGBA{30, 38, 50, 255},
			qrBg:   color.RGBA{255, 255, 255, 255},
		}
	default: // cyber
		return passTheme{
			bg:     color.RGBA{12, 14, 26, 255},
			header: color.RGBA{24, 10, 48, 255},
			border: color.RGBA{120, 0, 160, 255},
			title:  color.RGBA{0, 255, 240, 255},
			text:   color.RGBA{235, 235, 255, 255},
			muted:  color.RGBA{150, 150, 190, 255},
			accent: color.RGBA{255, 0, 180, 255},
			danger: color.RGBA{255, 80, 120, 255},
			qrFg:   color.RGBA{10, 10, 20, 255},
			qrBg:   color.RGBA{255, 255, 255, 255},
		}
	}
}

// passFace — TTF-шрифт с поддержкой кириллицы (Go Regular).
var passFace font.Face

func init() {
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		log.Printf("parse font: %v", err)
		return
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: 15, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		log.Printf("new face: %v", err)
		return
	}
	passFace = face
}

func GeneratePassImage(p Profile, theme string) ([]byte, error) {
	t := themeColors(theme)

	width, height := 660, 420
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.Draw(img, img.Bounds(), &image.Uniform{t.bg}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, 0, width, 64), &image.Uniform{t.header}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, 62, width, 66), &image.Uniform{t.accent}, image.Point{}, draw.Src)

	borderColor := t.border
	for x := 0; x < width; x++ {
		img.Set(x, 0, borderColor)
		img.Set(x, height-1, borderColor)
	}
	for y := 0; y < height; y++ {
		img.Set(0, y, borderColor)
		img.Set(width-1, y, borderColor)
	}

	// Загружаем аватар с заголовком User-Agent
	if p.AvatarURL != "" {
		client := &http.Client{Timeout: 7 * time.Second}
		req, err := http.NewRequest("GET", p.AvatarURL, nil)
		if err == nil {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				avatarImg, _, err := image.Decode(resp.Body)
				if err == nil {
					avatarRect := image.Rect(28, 92, 188, 252)
					draw.ApproxBiLinear.Scale(img, avatarRect, avatarImg, avatarImg.Bounds(), draw.Over, nil)
				} else {
					log.Printf("ошибка декодирования аватара: %v", err)
				}
			}
		}
	}

	drawOutline(img, image.Rect(26, 90, 190, 254), t.accent, 2)

	// Заголовок
	addText(img, 24, 40, "SECURITY PASS", t.title)
	addText(img, 470, 40, "ID: "+p.Passport, t.title)

	// Данные владельца — кириллица теперь отображается корректно
	addText(img, 212, 118, "ФИО: "+p.LastName+" "+p.FirstName, t.text)
	addText(img, 212, 150, "ГОРОД: "+p.City, t.muted)
	addText(img, 212, 182, "ПОЗЫВНОЙ: "+p.Callsign, t.accent)
	addText(img, 212, 214, "ГРУППА КРОВИ: "+p.Blood, t.danger)
	addText(img, 212, 246, "ДОСТУП: LEVEL 3 (RESTRICTED)", t.text)

	// Нижний блок: паспорт и QR-код
	addText(img, 28, 300, "ПАСПОРТ: "+p.Passport, t.muted)
	addText(img, 28, 328, "ВЫДАН: "+time.Now().Format("02.01.2006"), t.muted)
	addText(img, 28, 386, "Служба безопасности", t.accent)

	payload, err := json.Marshal(map[string]string{
		"passport": p.Passport,
		"name":     p.LastName + " " + p.FirstName,
		"city":     p.City,
		"callsign": p.Callsign,
		"blood":    p.Blood,
	})
	if err == nil {
		drawQRCode(img, 500, 262, 138, string(payload), t.qrFg, t.qrBg)
		drawOutline(img, image.Rect(490, 252, 648, 410), t.accent, 2)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawOutline(img *image.RGBA, rect image.Rectangle, col color.Color, thickness int) {
	for t := 0; t < thickness; t++ {
		for x := rect.Min.X - t; x <= rect.Max.X+t; x++ {
			img.Set(x, rect.Min.Y-t, col)
			img.Set(x, rect.Max.Y+t, col)
		}
		for y := rect.Min.Y - t; y <= rect.Max.Y+t; y++ {
			img.Set(rect.Min.X-t, y, col)
			img.Set(rect.Max.X+t, y, col)
		}
	}
}

func addText(img *image.RGBA, x, y int, label string, col color.Color) {
	face := passFace
	if face == nil {
		face = basicfont.Face7x13
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(label)
}

// drawQRCode рисует настоящий QR-код с данными профиля.
// x, y, size задают область самих модулей; вокруг добавляется quiet-zone.
func drawQRCode(img *image.RGBA, x, y, size int, content string, fg, bg color.Color) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		log.Printf("qrcode: %v", err)
		return
	}

	const pad = 8
	draw.Draw(img, image.Rect(x-pad, y-pad, x+size+pad, y+size+pad), &image.Uniform{bg}, image.Point{}, draw.Src)

	bitmap := qr.Bitmap()
	n := len(bitmap)
	if n == 0 {
		return
	}

	cell := float64(size) / float64(n)
	for row := 0; row < n; row++ {
		for col := 0; col < n; col++ {
			if !bitmap[row][col] {
				continue
			}
			x0 := x + int(float64(col)*cell)
			y0 := y + int(float64(row)*cell)
			x1 := x + int(float64(col+1)*cell)
			y1 := y + int(float64(row+1)*cell)
			if x1 <= x0 {
				x1 = x0 + 1
			}
			if y1 <= y0 {
				y1 = y0 + 1
			}
			draw.Draw(img, image.Rect(x0, y0, x1, y1), &image.Uniform{fg}, image.Point{}, draw.Src)
		}
	}
}
