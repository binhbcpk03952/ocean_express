const fs = require('fs');

const hubsData = fs.readFileSync('hubs.json', 'utf16le').trim();
// hubs.json might have multiple lines if formatted, but psql -A outputs an array or something.
// Wait, psql json_agg outputs a single line JSON array.
let hubs = [];
try {
  hubs = JSON.parse(hubsData);
} catch (e) {
  console.error("Failed to parse hubs.json", e);
  process.exit(1);
}

const passwordHash = '$2a$14$2MR2EoYmv963tCyU2wWT7e4Ru5c1byA.cQEOkYY6AxQYixZt9xUW2'; // From admin 0900000000
let phoneCounter = 910000001;

function toSlug(str) {
  if (!str) return 'hub';
  str = str.toLowerCase();
  str = str.replace(/à|á|ạ|ả|ã|â|ầ|ấ|ậ|ẩ|ẫ|ă|ằ|ắ|ặ|ẳ|ẵ/g, "a");
  str = str.replace(/è|é|ẹ|ẻ|ẽ|ê|ề|ế|ệ|ể|ễ/g, "e");
  str = str.replace(/ì|í|ị|ỉ|ĩ/g, "i");
  str = str.replace(/ò|ó|ọ|ỏ|õ|ô|ồ|ố|ộ|ổ|ỗ|ơ|ờ|ớ|ợ|ở|ỡ/g, "o");
  str = str.replace(/ù|ú|ụ|ủ|ũ|ư|ừ|ứ|ự|ử|ữ/g, "u");
  str = str.replace(/ỳ|ý|ỵ|ỷ|ỹ/g, "y");
  str = str.replace(/đ/g, "d");
  str = str.replace(/\s+/g, "_");
  str = str.replace(/[^a-z0-9_]/g, "");
  return str;
}

const roles = [
  { code: 'hub_staff', name: 'Nhân viên kho' },
  { code: 'first_mile_driver', name: 'Tài xế nhận hàng' },
  { code: 'last_mile_driver', name: 'Tài xế giao hàng' }
];

let sql = `-- Seeding employees for all hubs\n`;
sql += `INSERT INTO employees (name, phone, email, password_hash, role, hub_id, status)\nVALUES\n`;

const values = [];

for (const hub of hubs) {
  const hubSlug = toSlug(hub.name);
  
  for (const role of roles) {
    // For SOC, maybe skip last_mile_driver? Actually, no harm in having them.
    if (hub.type === 'soc' && role.code === 'last_mile_driver') continue;
    
    const empName = `${role.name} - ${hub.name}`;
    const phone = `0${phoneCounter++}`;
    const email = `${role.code}_${hubSlug}@oceanexpress.com`;
    
    values.push(`('${empName}', '${phone}', '${email}', '${passwordHash}', '${role.code}', '${hub.id}', 'approved')`);
  }
}

sql += values.join(',\n') + ';\n';

fs.writeFileSync('seed_employees.sql', sql);
console.log(`Generated ${values.length} employees.`);
