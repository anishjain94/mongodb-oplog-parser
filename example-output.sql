CREATE SCHEMA IF NOT EXISTS student;

CREATE TABLE IF NOT EXISTS student.students(_id VARCHAR(255) PRIMARY KEY,age FLOAT,name VARCHAR(255),subject VARCHAR(255));

INSERT INTO student.students(name, subject, _id, age) VALUES ('Nathan Lindgren', 'Maths', '64798c213f273a7ca2cf516a', 25.000000);

INSERT INTO student.students(_id, age, name, subject) VALUES ('64798c213f273a7ca2cf516b', 18.000000, 'Meggie Hoppe', 'English');

CREATE SCHEMA IF NOT EXISTS employee;

CREATE TABLE IF NOT EXISTS employee.employees_phone(employees__id VARCHAR(255),personal VARCHAR(255),work VARCHAR(255),_id VARCHAR(255) PRIMARY KEY);

INSERT INTO employee.employees_phone(personal, work, _id, employees__id) VALUES ('8764255212', '2762135091', 'e1916f3d-c2bf-44a6-9966-516c16442a8c', '64798c213f273a7ca2cf516c');

CREATE TABLE IF NOT EXISTS employee.employees_address(line1 VARCHAR(255),zip VARCHAR(255),_id VARCHAR(255) PRIMARY KEY,employees__id VARCHAR(255));

INSERT INTO employee.employees_address(line1, zip, _id, employees__id) VALUES ('32550 Port Gatewaytown', '18399', '606fc9b9-10ee-4caf-aa0c-84eb1c341a2d', '64798c213f273a7ca2cf516c');

INSERT INTO employee.employees_address(line1, zip, _id, employees__id) VALUES ('3840 Cornermouth', '83941', '5adc5aa5-f260-4e05-ae34-3bdb5bf04a67', '64798c213f273a7ca2cf516c');

CREATE TABLE IF NOT EXISTS employee.employees(salary FLOAT,_id VARCHAR(255) PRIMARY KEY,age FLOAT,name VARCHAR(255),position VARCHAR(255));

INSERT INTO employee.employees(salary, _id, age, name, position) VALUES (3767.925635, '64798c213f273a7ca2cf516c', 35.000000, 'Raymond Monahan', 'Engineer');

DELETE FROM student.students WHERE _id = '64798c213f273a7ca2cf516a';

INSERT INTO student.students(subject, _id, age, name) VALUES ('English', '64798c213f273a7ca2cf516d', 19.000000, 'Tevin Heathcote');

INSERT INTO employee.employees_phone(personal, work, _id, employees__id) VALUES ('7678456640', '8130097989', 'cd8bbd12-2b76-4955-8d82-5e1e8e80a30a', '64798c213f273a7ca2cf516e');

INSERT INTO employee.employees_address(line1, zip, _id, employees__id) VALUES ('481 Harborsburgh', '89799', '5db80483-9a51-4293-92ff-09be8981df2a', '64798c213f273a7ca2cf516e');

INSERT INTO employee.employees_address(employees__id, line1, zip, _id) VALUES ('64798c213f273a7ca2cf516e', '329 Flatside', '80872', '2c451fbd-04ed-409f-8f7c-bb98feeba6bf');

INSERT INTO employee.employees(age, name, position, salary, _id) VALUES (37.000000, 'Wilson Gleason', 'Manager', 5042.121824, '64798c213f273a7ca2cf516e');

INSERT INTO employee.employees_phone(_id, employees__id, personal, work) VALUES ('36667a70-181b-4596-b0f1-84ad5e9fa0ba', '64798c213f273a7ca2cf516f', '1075027422', '1641587035');

INSERT INTO employee.employees_address(employees__id, line1, zip, _id) VALUES ('64798c213f273a7ca2cf516f', '96400 Landhaven', '41638', 'b472dc10-6da6-46f6-92a9-6bd6d542231c');

INSERT INTO employee.employees_address(line1, zip, _id, employees__id) VALUES ('3939 Lightburgh', '99747', 'ae0b6498-8bbb-430c-9e57-8ce03b6467d8', '64798c213f273a7ca2cf516f');

INSERT INTO employee.employees(name, position, salary, _id, age) VALUES ('Linwood Wilkinson', 'Manager', 4514.763474, '64798c213f273a7ca2cf516f', 31.000000);

INSERT INTO student.students(age, name, subject, _id) VALUES (18.000000, 'Camren Thompson', 'Science', '64798c213f273a7ca2cf5170');

INSERT INTO employee.employees_address(line1, zip, _id, employees__id) VALUES ('51338 Landingbury', '74795', '5ab5120a-b661-4efb-bc43-f9053d829a86', '64798c213f273a7ca2cf5171');

INSERT INTO employee.employees_address(line1, zip, _id, employees__id) VALUES ('79033 West Locksmouth', '43555', '2bae6244-2218-4f91-bfd5-8082b3ccbd60', '64798c213f273a7ca2cf5171');

INSERT INTO employee.employees_phone(_id, employees__id, personal, work) VALUES ('8e357e55-a8de-4fc5-bcc8-28008a7b1211', '64798c213f273a7ca2cf5171', '4613562303', '1889316722');

INSERT INTO employee.employees(position, salary, _id, age, name) VALUES ('Engineer', 6676.956104, '64798c213f273a7ca2cf5171', 31.000000, 'Meaghan Hettinger');

UPDATE employee.employees SET Age = 23.000000 WHERE _id = '64798c213f273a7ca2cf5171';

INSERT INTO employee.employees_address(_id, employees__id, line1, zip) VALUES ('92c51b1e-7893-4e22-a708-016635953bba', '64798c213f273a7ca2cf5172', '2787 Trackview', '23598');

INSERT INTO employee.employees_address(line1, zip, _id, employees__id) VALUES ('33659 South Mountainchester', '45086', '2e7e0439-1b14-46c8-9c96-5382f5bb4f2a', '64798c213f273a7ca2cf5172');

INSERT INTO employee.employees_phone(employees__id, personal, work, _id) VALUES ('64798c213f273a7ca2cf5172', '9829848796', '5636590993', '9b02a8c9-db52-4df2-85e6-7a07adf2d182');

ALTER TABLE employee.employees ADD workhours FLOAT;

INSERT INTO employee.employees(age, name, position, salary, workhours, _id) VALUES (20.000000, 'Delta Bahringer', 'Developer', 2980.127110, 6.000000, '64798c213f273a7ca2cf5172');

ALTER TABLE student.students ADD is_graduated BOOLEAN;

INSERT INTO student.students(_id, age, is_graduated, name, subject) VALUES ('64798c213f273a7ca2cf5173', 20.000000, false, 'Freda Dare', 'Maths');

INSERT INTO student.students(_id, age, is_graduated, name, subject) VALUES ('64798c213f273a7ca2cf5174', 23.000000, true, 'Kamille Jast', 'Maths');

INSERT INTO student.students(age, is_graduated, name, subject, _id) VALUES (19.000000, false, 'Arden Kessler', 'Social Studies', '64798c213f273a7ca2cf5175');

INSERT INTO employee.employees_address(_id, employees__id, line1, zip) VALUES ('a5cde101-4fa3-4740-bf9b-c7cddd54a840', '64798c213f273a7ca2cf5176', '403 Walksfurt', '75756');

INSERT INTO employee.employees_address(_id, employees__id, line1, zip) VALUES ('b3eb307c-7e1e-43c1-836b-eba86683f959', '64798c213f273a7ca2cf5176', '5012 Port Branchberg', '21969');

INSERT INTO employee.employees_phone(work, _id, employees__id, personal) VALUES ('2515301788', 'ee8655c6-aa8a-481c-9082-17923ac8b615', '64798c213f273a7ca2cf5176', '1748534264');

INSERT INTO employee.employees(position, salary, workhours, _id, age, name) VALUES ('Salesman', 6322.655858, 4.000000, '64798c213f273a7ca2cf5176', 29.000000, 'Chyna Kihn');

INSERT INTO employee.employees_address(_id, employees__id, line1, zip) VALUES ('8a6c640e-6c29-4cac-b0ea-6b2b630ef8de', '64798c213f273a7ca2cf5177', '73628 Port Knollchester', '97436');

INSERT INTO employee.employees_address(line1, zip, _id, employees__id) VALUES ('93072 Lake Skywayhaven', '87218', '37245837-00f1-4329-b075-ecdd59da5387', '64798c213f273a7ca2cf5177');

INSERT INTO employee.employees_phone(work, _id, employees__id, personal) VALUES ('9172896730', 'f3361cf6-bf1a-4c98-9318-eb1b811517c2', '64798c213f273a7ca2cf5177', '1498807115');

INSERT INTO employee.employees(position, salary, workhours, _id, age, name) VALUES ('Engineer', 9811.365188, 5.000000, '64798c213f273a7ca2cf5177', 38.000000, 'Madie Klein');

