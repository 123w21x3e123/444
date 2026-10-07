(()=>{
const $=(s,r=document)=>r.querySelector(s);
const esc=s=>String(s).replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]));
function toast(m){let t=$('#c444-toast');if(!t){t=document.createElement('div');t.id='c444-toast';document.body.appendChild(t)}
 t.textContent=m;t.style.opacity=1;clearTimeout(t._h);t._h=setTimeout(()=>t.style.opacity=0,2500)}

/* ---- hide picker: Alt+Click hides an element, Alt+Shift+Z undoes ---- */
const KEY='444.hidden';let hidden=[];
try{hidden=JSON.parse(localStorage.getItem(KEY)||'[]')}catch(e){}
const hs=document.createElement('style');document.head.appendChild(hs);
const saveHidden=()=>{hs.textContent=hidden.map(s=>s+'{display:none!important}').join('\n');try{localStorage.setItem(KEY,JSON.stringify(hidden))}catch(e){}};
saveHidden();
function sel(el){
 const path=[];
 while(el&&el.nodeType===1&&el!==document.body){
  let s=el.tagName.toLowerCase();
  const t=el.getAttribute('data-testid'),a=el.getAttribute('aria-label');
  if(el.id){path.unshift('#'+CSS.escape(el.id));break}
  if(t){path.unshift(s+'[data-testid="'+t+'"]');break}
  if(a)s+='[aria-label="'+a.replace(/"/g,'\\"')+'"]';
  else s+=':nth-child('+([...el.parentNode.children].indexOf(el)+1)+')';
  path.unshift(s);el=el.parentElement;
 }
 return path.join('>');
}
addEventListener('click',e=>{
 if(!e.altKey||e.shiftKey)return;
 e.preventDefault();e.stopPropagation();
 const s=sel(e.target);if(!s)return;
 hidden.push(s);saveHidden();
 toast('Hidden '+document.querySelectorAll(s).length+' item(s). Alt+Shift+Z to undo');
},true);

/* ---- lyrics overlay: Alt+L, Esc to close (synced lyrics from lrclib.net) ---- */
const secs=t=>{const p=(t||'').split(':').map(Number).reverse();return(p[0]||0)+(p[1]||0)*60+(p[2]||0)*3600};
const txt=s=>{const e=$(s);return e?e.textContent.trim():''};
const track=()=>({title:txt('[data-testid="context-item-link"]'),artist:txt('[data-testid="context-item-info-artist"]'),
 pos:secs(txt('[data-testid="playback-position"]')),dur:secs(txt('[data-testid="playback-duration"]')),
 cover:($('[data-testid="cover-art-image"]')||{}).src||''});
const ov=document.createElement('div');ov.id='c444-ly';
ov.innerHTML='<div class="bg"></div><div class="hd"><b></b><span></span></div><div class="ln"></div>';
document.body.appendChild(ov);
const ln=$('.ln',ov);
let lines=[],key='',loaded='',cur=-1,open=false,last=-1,t0=0;
async function load(t){
 loaded=key;lines=[];cur=-1;ln.innerHTML='<p class="m">Loading lyrics…</p>';
 try{
  const q=new URLSearchParams({track_name:t.title,artist_name:t.artist});
  const j=await(await fetch('https://lrclib.net/api/search?'+q)).json();
  const syn=j.filter(x=>x.syncedLyrics).sort((a,b)=>Math.abs(a.duration-t.dur)-Math.abs(b.duration-t.dur))[0];
  if(syn){
   lines=syn.syncedLyrics.split('\n').map(l=>{const m=l.match(/^\[(\d+):(\d+(?:\.\d+)?)\](.*)/);
    return m?{t:Number(m[1])*60+Number(m[2]),x:m[3].trim()}:null}).filter(Boolean);
   ln.innerHTML=lines.map(l=>'<p>'+(esc(l.x)||'♪')+'</p>').join('');
  }else if(j[0]&&j[0].plainLyrics){
   ln.innerHTML=j[0].plainLyrics.split('\n').map(x=>'<p class="plain">'+esc(x)+'</p>').join('');
  }else ln.innerHTML='<p class="m">No lyrics found.</p>';
 }catch(e){ln.innerHTML='<p class="m">Couldn\'t load lyrics ('+esc(e.message)+').</p>'}
}
function show(on){
 open=on;ov.classList.toggle('open',on);
 if(on){const t=track();$('.bg',ov).style.backgroundImage=t.cover?'url("'+t.cover+'")':'none';
  $('.hd b',ov).textContent=t.title;$('.hd span',ov).textContent=t.artist;
  if(t.title&&loaded!==t.title+'|'+t.artist)load(t)}
}
addEventListener('keydown',e=>{
 if(e.altKey&&e.shiftKey&&e.code==='KeyZ'){hidden.pop();saveHidden();toast('Undone')}
 else if(e.altKey&&!e.shiftKey&&e.code==='KeyL'){e.preventDefault();show(!open)}
 else if(e.code==='Escape'&&open)show(false);
});
ov.onclick=()=>show(false);
setInterval(()=>{
 const t=track();if(!t.title)return;
 const k=t.title+'|'+t.artist;
 if(k!==key){key=k;if(open)show(true)}
 if(!open||!lines.length)return;
 if(t.pos!==last){last=t.pos;t0=performance.now()}
 const est=t.pos+Math.min((performance.now()-t0)/1000,.99);
 let i=-1;for(let j=0;j<lines.length;j++){if(lines[j].t<=est+.15)i=j;else break}
 if(i!==cur){cur=i;[...ln.children].forEach((p,j)=>p.className=j===i?'on':j<i?'past':'');
  if(ln.children[i])ln.children[i].scrollIntoView({block:'center',behavior:'smooth'})}
},250);
})();
