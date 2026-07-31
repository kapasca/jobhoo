package main

import "github.com/jobhoo/jobhoo/internal/models"

type categorySpec struct {
	Value      models.JobCategory
	Industry   string
	Titles     []string
	MustHave   []string
	NiceToHave []string
}

// jobCategories seeds exactly the 5 categories JOBHOO classifies jobs into.
// Each carries its own title/skill pool so generated jobs read as plausible
// rather than randomly assembled words.
var jobCategories = []categorySpec{
	{
		Value:      models.CategoryEngineeringProduct,
		Industry:   "Technology",
		Titles:     []string{"Backend Engineer", "Frontend Engineer", "Full Stack Engineer", "Mobile Engineer", "DevOps Engineer", "Site Reliability Engineer", "Product Manager", "QA Engineer", "Platform Engineer", "Engineering Manager"},
		MustHave:   []string{"Go", "JavaScript", "Python", "React", "SQL", "Kubernetes", "AWS", "TypeScript", "Java", "REST APIs"},
		NiceToHave: []string{"Docker", "GraphQL", "Terraform", "CI/CD", "gRPC", "PostgreSQL", "Redis", "Kafka"},
	},
	{
		Value:      models.CategoryDesignCreative,
		Industry:   "Design",
		Titles:     []string{"Product Designer", "UX Researcher", "UI Designer", "Brand Designer", "Motion Designer", "Design Systems Lead", "Creative Director", "Illustrator"},
		MustHave:   []string{"Figma", "Design Systems", "Prototyping", "User Research", "Typography", "Interaction Design"},
		NiceToHave: []string{"After Effects", "Illustrator", "Webflow", "Accessibility", "Design Ops"},
	},
	{
		Value:      models.CategorySalesMarketing,
		Industry:   "Sales & Marketing",
		Titles:     []string{"Account Executive", "SDR", "Growth Marketer", "Content Marketing Manager", "Marketing Ops Manager", "Partnerships Manager", "Sales Development Rep", "Demand Generation Manager"},
		MustHave:   []string{"CRM", "Salesforce", "Negotiation", "Lead Generation", "Copywriting", "SEO"},
		NiceToHave: []string{"HubSpot", "Paid Media", "Email Marketing", "Account-Based Marketing", "Analytics"},
	},
	{
		Value:      models.CategoryDataAnalytics,
		Industry:   "Data",
		Titles:     []string{"Data Analyst", "Data Scientist", "Data Engineer", "Analytics Engineer", "BI Developer", "Machine Learning Engineer", "Data Platform Lead"},
		MustHave:   []string{"SQL", "Python", "Data Modeling", "Statistics", "ETL", "dbt"},
		NiceToHave: []string{"Looker", "Airflow", "Spark", "Tableau", "A/B Testing", "Snowflake"},
	},
	{
		Value:      models.CategoryOperationsSupport,
		Industry:   "Operations",
		Titles:     []string{"Operations Manager", "Customer Support Specialist", "People Operations Manager", "Recruiting Coordinator", "Office Manager", "Supply Chain Analyst", "Customer Success Manager"},
		MustHave:   []string{"Process Improvement", "Stakeholder Management", "Communication", "Project Management"},
		NiceToHave: []string{"Zendesk", "Notion", "Airtable", "Vendor Management"},
	},
}

var companyNames = []struct {
	Name     string
	Industry string
}{
	{"Acme Robotics", "Robotics"},
	{"Northwind Data", "Analytics"},
	{"Bluepeak Software", "Technology"},
	{"Harborlight Studio", "Design"},
	{"Everline Logistics", "Logistics"},
	{"Solace Health", "Healthcare"},
	{"Fernwood Retail", "Retail"},
	{"Ironclad Financial", "Finance"},
	{"Meridian Labs", "Biotech"},
	{"Cobalt Media", "Media"},
}

type seedLocation struct {
	Country string
	State   string
}

var seedLocations = []seedLocation{
	{"Indonesia", "DKI Jakarta"},
	{"Indonesia", "Jawa Barat"},
	{"Indonesia", "Jawa Timur"},
	{"Indonesia", "Bali"},
	{"Indonesia", "Jawa Tengah"},
	{"Indonesia", "Sumatera Utara"},
	{"Singapore", "Singapore"},
	{"Malaysia", "Kuala Lumpur"},
	{"Malaysia", "Selangor"},
	{"Australia", "New South Wales"},
	{"Australia", "Victoria"},
	{"Thailand", "Bangkok"},
	{"Vietnam", "Hà Nội"},
	{"Vietnam", "TP. Hồ Chí Minh"},
}

var seniorities = []string{"Junior", "Mid", "Senior", "Lead"}
