package models

// CountryEntry pairs a country name with its administrative subdivisions.
type CountryEntry struct {
	Name     string
	States   []string // province / state / region / municipality depending on country
	CityOnly bool     // true for Singapore — auto-fill state and disable the dropdown
	Currency string   // ISO 4217 currency code used as default for jobs in this country
}

// Countries is the ordered list of countries JOBHOO supports for job location,
// with Indonesia first as the primary market.
var Countries = []CountryEntry{
	{
		Name:     "Indonesia",
		Currency: "IDR",
		States: []string{
			"Aceh", "Bali", "Bangka Belitung", "Banten", "Bengkulu",
			"DI Yogyakarta", "DKI Jakarta", "Gorontalo", "Jambi",
			"Jawa Barat", "Jawa Tengah", "Jawa Timur",
			"Kalimantan Barat", "Kalimantan Selatan", "Kalimantan Tengah",
			"Kalimantan Timur", "Kalimantan Utara", "Kepulauan Riau",
			"Lampung", "Maluku", "Maluku Utara",
			"Nusa Tenggara Barat", "Nusa Tenggara Timur",
			"Papua", "Papua Barat", "Papua Barat Daya", "Papua Pegunungan",
			"Papua Selatan", "Papua Tengah", "Riau",
			"Sulawesi Barat", "Sulawesi Selatan", "Sulawesi Tengah",
			"Sulawesi Tenggara", "Sulawesi Utara",
			"Sumatera Barat", "Sumatera Selatan", "Sumatera Utara",
		},
	},
	{
		Name:     "Singapore",
		Currency: "SGD",
		States:   []string{"Singapore"},
		CityOnly: true,
	},
	{
		Name:     "Malaysia",
		Currency: "MYR",
		States: []string{
			"Johor", "Kedah", "Kelantan", "Melaka", "Negeri Sembilan",
			"Pahang", "Perak", "Perlis", "Pulau Pinang", "Sabah",
			"Sarawak", "Selangor", "Terengganu",
			"Kuala Lumpur", "Labuan", "Putrajaya",
		},
	},
	{
		Name:     "Thailand",
		Currency: "THB",
		States: []string{
			"Amnat Charoen", "Ang Thong", "Bangkok", "Bueng Kan",
			"Buri Ram", "Chachoengsao", "Chai Nat", "Chaiyaphum",
			"Chanthaburi", "Chiang Mai", "Chiang Rai", "Chon Buri",
			"Chumphon", "Kalasin", "Kamphaeng Phet", "Kanchanaburi",
			"Khon Kaen", "Krabi", "Lampang", "Lamphun", "Loei",
			"Lop Buri", "Mae Hong Son", "Maha Sarakham", "Mukdahan",
			"Nakhon Nayok", "Nakhon Pathom", "Nakhon Phanom",
			"Nakhon Ratchasima", "Nakhon Sawan", "Nakhon Si Thammarat",
			"Nan", "Narathiwat", "Nong Bua Lam Phu", "Nong Khai",
			"Nonthaburi", "Pathum Thani", "Pattani", "Phangnga",
			"Phatthalung", "Phayao", "Phetchabun", "Phetchaburi",
			"Phichit", "Phitsanulok", "Phra Nakhon Si Ayutthaya",
			"Phrae", "Phuket", "Prachin Buri", "Prachuap Khiri Khan",
			"Ranong", "Ratchaburi", "Rayong", "Roi Et", "Sa Kaeo",
			"Sakon Nakhon", "Samut Prakan", "Samut Sakhon",
			"Samut Songkhram", "Saraburi", "Satun", "Sing Buri",
			"Si Sa Ket", "Songkhla", "Sukhothai", "Suphan Buri",
			"Surat Thani", "Surin", "Tak", "Trang", "Trat",
			"Ubon Ratchathani", "Udon Thani", "Uthai Thani",
			"Uttaradit", "Yala", "Yasothon",
		},
	},
	{
		Name:     "Vietnam",
		Currency: "VND",
		States: []string{
			"An Giang", "Bà Rịa–Vũng Tàu", "Bắc Giang", "Bắc Kạn",
			"Bạc Liêu", "Bắc Ninh", "Bến Tre", "Bình Định",
			"Bình Dương", "Bình Phước", "Bình Thuận", "Cà Mau",
			"Cần Thơ", "Cao Bằng", "Đà Nẵng", "Đắk Lắk", "Đắk Nông",
			"Điện Biên", "Đồng Nai", "Đồng Tháp", "Gia Lai",
			"Hà Giang", "Hà Nam", "Hà Nội", "Hà Tĩnh", "Hải Dương",
			"Hải Phòng", "Hậu Giang", "Hòa Bình", "Hưng Yên",
			"Khánh Hòa", "Kiên Giang", "Kon Tum", "Lai Châu",
			"Lâm Đồng", "Lạng Sơn", "Lào Cai", "Long An",
			"Nam Định", "Nghệ An", "Ninh Bình", "Ninh Thuận",
			"Phú Thọ", "Phú Yên", "Quảng Bình", "Quảng Nam",
			"Quảng Ngãi", "Quảng Ninh", "Quảng Trị", "Sóc Trăng",
			"Sơn La", "Tây Ninh", "Thái Bình", "Thái Nguyên",
			"Thanh Hóa", "Thừa Thiên Huế", "Tiền Giang",
			"TP. Hồ Chí Minh", "Trà Vinh", "Tuyên Quang",
			"Vĩnh Long", "Vĩnh Phúc", "Yên Bái",
		},
	},
	{
		Name:     "Philippines",
		Currency: "PHP",
		States: []string{
			"NCR (Metro Manila)", "CAR",
			"Ilocos Region", "Cagayan Valley", "Central Luzon",
			"CALABARZON", "MIMAROPA", "Bicol Region",
			"Western Visayas", "Central Visayas", "Eastern Visayas",
			"Zamboanga Peninsula", "Northern Mindanao", "Davao Region",
			"SOCCSKSARGEN", "Caraga", "BARMM",
		},
	},
	{
		Name:     "Timor-Leste",
		Currency: "USD",
		States: []string{
			"Aileu", "Ainaro", "Baucau", "Bobonaro", "Cova Lima",
			"Dili", "Ermera", "Lautem", "Liquiçá", "Manatuto",
			"Manufahi", "Oecusse", "Viqueque",
		},
	},
	{
		Name:     "Australia",
		Currency: "AUD",
		States: []string{
			"New South Wales", "Victoria", "Queensland",
			"Western Australia", "South Australia", "Tasmania",
			"Australian Capital Territory", "Northern Territory",
		},
	},
	{
		Name:     "New Zealand",
		Currency: "NZD",
		States: []string{
			"Northland", "Auckland", "Waikato", "Bay of Plenty",
			"Gisborne", "Hawke's Bay", "Taranaki",
			"Manawatū-Whanganui", "Wellington", "Tasman", "Nelson",
			"Marlborough", "West Coast", "Canterbury", "Otago", "Southland",
		},
	},
}
