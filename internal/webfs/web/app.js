"use strict";
// Frontend: native HTML/CSS/JS. Calls the real business API; no framework.
const api = (p, o={}) => fetch(p, Object.assign({headers:{'Content-Type':'application/json'}}, o))
  .then(async r => { const t = await r.text().catch(()=> ''); let j=null; try{ j=t?JSON.parse(t):null }catch(e){} return {ok:r.ok, status:r.status, body:j, text:t} });

let currentProject = null;
let currentComponent = null;

const $ = id => document.getElementById(id);

async function loadProjects(){
  const {body} = await api('/api/projects');
  const ul = $('project-list'); ul.innerHTML='';
  (body||[]).forEach(p=>{
    const li = document.createElement('li');
    li.textContent = p.code + ' — ' + p.name + ' (V=' + p.wind_speed/10 + 'm/s, 暴露'+p.exposure+', 重要性'+p.importance+')';
    li.onclick = () => openProject(p.id);
    ul.appendChild(li);
  });
}

async function openProject(id){
  currentProject = id;
  $('detail').hidden = false;
  const {body:p} = await api('/api/projects/'+id);
  $('detail-title').textContent = p.code + ' / ' + p.name;
  await loadLevels(); await loadComponents(); await loadCombos(); await loadChecks(); await loadSummary();
}

async function loadLevels(){
  const {body} = await api('/api/projects/'+currentProject+'/levels');
  const ul=$('level-list'); ul.innerHTML='';
  (body||[]).forEach(l=>{ const li=document.createElement('li'); li.textContent=l.name+' 标高'+l.elevation+'mm 面'+l.gross_area+'cm²'; ul.appendChild(li); });
}

async function loadComponents(){
  const {body} = await api('/api/projects/'+currentProject+'/components');
  const ul=$('component-list'); ul.innerHTML='';
  (body||[]).forEach(c=>{
    const li=document.createElement('li'); li.dataset.id=c.id;
    li.textContent=c.code+' ('+c.type+', L='+c.span+'mm, Mn='+c.nominal_moment+')';
    li.onclick=()=>{ currentComponent=c.id; document.querySelectorAll('#component-list li').forEach(x=>x.classList.remove('active')); li.classList.add('active'); $('pick-component-lc').disabled=false; loadLoadCases(c.id); loadDerived(c.id); };
    ul.appendChild(li);
  });
}

async function loadLoadCases(cid){
  const {body} = await api('/api/components/'+cid+'/loadcases');
  const ul=$('loadcase-list'); ul.innerHTML='';
  (body||[]).forEach(lc=>{ const li=document.createElement('li'); li.textContent=lc.kind+' '+lc.load_type+' ='+lc.magnitude+' ['+lc.direction+'] ('+lc.origin+')'; ul.appendChild(li); });
}
async function loadDerived(cid){
  const {body} = await api('/api/components/'+cid+'/derived');
  body&&body.forEach(()=>{}); // derived listed in loadcase list already
}

async function loadCombos(){
  const {body} = await api('/api/projects/'+currentProject+'/combinations');
  const ul=$('combo-list'); ul.innerHTML='';
  (body||[]).forEach(c=>{ const li=document.createElement('li'); li.textContent=c.name+' D'+c.coeff_d+'/L'+c.coeff_l+'/S'+c.coeff_s+'/W'+c.coeff_w+' ['+c.kind+']'; ul.appendChild(li); });
}

async function loadChecks(){
  const {body} = await api('/api/projects/'+currentProject+'/checks');
  const tb=$('check-table').querySelector('tbody'); tb.innerHTML='';
  (body||[]).forEach(ch=>{
    const tr=document.createElement('tr');
    tr.innerHTML='<td>'+ch.component_id.slice(-4)+'</td><td>'+ch.combination_id.slice(-4)+'</td><td>'+ch.demand_moment+'</td><td>'+ch.capacity_moment+'</td><td>'+ch.ur+'</td><td class="status-'+ch.status+'">'+ch.status+'</td><td>'+ch.governing_kind+'</td>';
    tb.appendChild(tr);
  });
}

async function loadSummary(){
  const {body} = await api('/api/projects/'+currentProject+'/summary');
  $('summary-box').textContent = JSON.stringify(body, null, 2);
}

function dialog(form){
  const dlg=$('dlg'); const f=$('dlg-form'); f.innerHTML=form;
  return new Promise(res=>{ f.onsubmit=()=>{ const o={}; new FormData(f).forEach((v,k)=>o[k]=v); res(o); dlg.close(); }; dlg.showModal(); });
}

$('new-project').onclick = async ()=>{
  const f = await dialog(`
    <label>编号 code</label><input name="code" value="B001">
    <label>名称 name</label><input name="name" value="示例办公楼">
    <label>暴露 B/C/D</label><input name="exposure" value="C">
    <label>重要性 I/II/III/IV</label><input name="importance" value="II">
    <label>风速 (m/s, 0.1步)</label><input name="wind_speed" value="400">
    <label>地面雪载 (0.01kPa)</label><input name="ground_snow" value="95">
    <button>创建</button>`);
  f.wind_speed=+f.wind_speed; f.ground_snow=+f.ground_snow;
  await api('/api/projects',{method:'POST',body:JSON.stringify(f)}); loadProjects();
};

$('add-level').onclick = async ()=>{
  const f = await dialog(`<label>名称</label><input name="name" value="1F"><label>标高 mm</label><input name="elevation" value="0"><label>面积 cm²</label><input name="gross_area" value="1000000"><button>加楼层</button>`);
  f.elevation=+f.elevation; f.gross_area=+f.gross_area;
  await api('/api/projects/'+currentProject+'/levels',{method:'POST',body:JSON.stringify(f)}); loadLevels();
};

$('add-component').onclick = async ()=>{
  // Components are added under a level; pick (or create) the first level.
  let lvls = (await api('/api/projects/'+currentProject+'/levels')).body||[];
  if (!lvls.length) {
    await api('/api/projects/'+currentProject+'/levels',{method:'POST',body:JSON.stringify({name:'1F',elevation:0,gross_area:1000000})});
    lvls = (await api('/api/projects/'+currentProject+'/levels')).body||[];
  }
  const levelID = lvls[0].id;
  const f = await dialog(`<label>编号</label><input name="code" value="B1"><label>类型 beam/column/slab</label><input name="type" value="beam"><label>跨度 mm</label><input name="span" value="6000"><label>从属面积 cm²</label><input name="tributary_area" value="400000"><label>从属宽度 mm</label><input name="tributary_width" value="2500"><label>KLL</label><input name="kll" value="2"><label>Mn 0.01kN·m</label><input name="nominal_moment" value="10000"><label>phi_b (centi)</label><input name="phi_b" value="90"><label>风面 none/windward/leeward/side/roof_uplift/net_mwfrs</label><input name="wind_surface" value="windward"><button>加构件</button>`);
  ['span','tributary_area','tributary_width','kll','nominal_moment','nominal_axial','nominal_shear','phi_b','phi_c','phi_v','roof_slope'].forEach(k=>f[k]=+f[k]||0);
  await api('/api/levels/'+levelID+'/components',{method:'POST',body:JSON.stringify(f)});
  loadComponents();
};

$('pick-component-lc').onclick = async ()=>{
  if(!currentComponent) return;
  const f = await dialog(`<label>类型 D/L/Lr/R/E</label><input name="kind" value="D"><label>载型 area_pressure/line_load/point_load</label><input name="load_type" value="area_pressure"><label>量值</label><input name="magnitude" value="150"><label>方向 gravity/uplift/lateral</label><input name="direction" value="gravity"><button>录工况</button>`);
  f.magnitude=+f.magnitude;
  await api('/api/components/'+currentComponent+'/loadcases',{method:'POST',body:JSON.stringify(f)}); loadLoadCases(currentComponent);
};

$('add-combo').onclick = async ()=>{
  const f = await dialog(`<label>名称</label><input name="name" value="1.2D+1.6L+0.5S"><label>类型 strength/service</label><input name="kind" value="strength"><label>coeff_d</label><input name="coeff_d" value="120"><label>coeff_l</label><input name="coeff_l" value="160"><label>coeff_s</label><input name="coeff_s" value="50"><button>建组合</button>`);
  ['coeff_d','coeff_l','coeff_lr','coeff_s','coeff_w','coeff_r','coeff_e'].forEach(k=>f[k]=+f[k]||0);
  await api('/api/projects/'+currentProject+'/combinations',{method:'POST',body:JSON.stringify(f)}); loadCombos();
};

$('derive').onclick = async ()=>{ await api('/api/projects/'+currentProject+'/derive',{method:'POST'}); if(currentComponent) loadLoadCases(currentComponent); };
$('run-checks').onclick = async ()=>{ await api('/api/projects/'+currentProject+'/checks/run',{method:'POST'}); loadChecks(); loadSummary(); };
$('recompute').onclick = async ()=>{ await api('/api/projects/'+currentProject+'/recompute',{method:'POST'}); loadChecks(); loadSummary(); };
$('audit').onclick = async ()=>{ const {body}=await api('/api/projects/'+currentProject+'/audit'); alert(JSON.stringify(body)); };
$('ovr-go').onclick = async ()=>{
  await api('/api/components/'+$('ovr-comp').value+'/override',{method:'POST',body:JSON.stringify({field:$('ovr-field').value,new_value:$('ovr-val').value,reason:$('ovr-reason').value})});
  loadComponents(); loadChecks(); loadSummary();
};

loadProjects();
